// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "./flat_deps.sol";
import "./IBEP20.sol";
import "./IPancakeRouter01.sol";

/**
 * @title P2E
 * @dev A contract for predicting and earning rewards based on the correct predictions, implemented as an ERC721 NFT token.
 */
contract P2E is ERC721, ERC721Enumerable, ERC721URIStorage, ERC721Burnable, Ownable {
    using Counters for Counters.Counter;

    // The prefix for the URL of prediction data
    string private constant urlPrefix = "example.com/prediction/";

    // Represents the last maximum finalized prediction NFT ID. Does not include NFT IDs of ongoing game.
    uint256 private _lastMaxNFT = 0;

    // The cost of minting a new prediction in FAME tokens, for current game.
    uint256 private _fameCostPerMint;

    // The FAME token used in the contract
    IBEP20 private _fame;

    // A mapping to store whether a given prediction hash exists. It prevents users from submitting duplicate predictions.
    mapping(bytes32 => bool) private _predictionExists;

    // A mapping to store the NFTs that won and their corresponding rewards
    mapping(uint256 => uint256) private _winNFTs;

    // The PancakeSwap Router used for swapping tokens when creating a prediction using an auxiliary token.
    IPancakeRouter01 private _swapRouter;

    // A mapping to store whether a given token address is whitelisted. Only whitelisted tokens can be used for creating predictions using auxiliary tokens.
    mapping(address => bool) private _whitelistedTokens;

    // An array to store the rewards percentages for each winning bracket (first, second, third place)
    uint256[3] private _bracketsRewardsPercentage;

    // The total rewards pool for the current round
    uint256 private _roundTotalPool = 0;

    // Guard variable that tracks how much FAME tokens are required by the contract, prevents owner to overwithdraw
    uint256 private _unspendableBalance = 0;

    // The percentage of the prediction cost to be added to the rewards pool
    uint256 private _priceToRewardPercentage = 0;

    // The percentage of the prediction cost to be used for consolidation rewards
    uint256 private _currentConsoliation = 0;

    // Enum to represent the current phase of the game
    enum Phase {MintingPhase, GamePhase, RewardsIdlePhase}
    Phase private _currentPhase = Phase.RewardsIdlePhase;

    // Counter for the NFT token IDs
    Counters.Counter private _tokenIdCounter;

    event NewPrediction(bytes32 indexed hash, uint256 nftId);
    event PhaseMintingStart();
    event PhaseGameStart();
    event PhaseIdleStart();
    event TokenWhitelistChange(address indexed token, bool permitted);

    constructor(IBEP20 fame, IPancakeRouter01 router) ERC721("Predict To Earn", "P2E") {
        _fame = fame;
        _swapRouter = router;
        _currentPhase = Phase.RewardsIdlePhase;
    }

    /**
    * @notice Allows the contract owner to withdraw the accumulated fees from the contract.
    * @dev When users create predictions, their tokens are allocated to reward and consolidation prize pools.
    *      Ratio is regulated by initRound params, if the sum of these percentages is below 100%, the remaining
    *      percentage of the tokens becomes a fee that accumulates in the contract. This function allows the
    *      contract owner to claim these fees as revenue.
    */
    function withdrawSpendable() onlyOwner public {
        uint256 bal = _fame.balanceOf(address(this));
        require(bal > _unspendableBalance, "insufficient funds");
        _fame.transfer(msg.sender, bal - _unspendableBalance);
    }

    /**
     * @notice Creates a prediction using an auxiliary token.
     * @dev The token must be whitelisted.
     * @param hash A unique hash representing the prediction.
     * @param token The auxiliary token address.
     */
    function createPredictionAuxToken(bytes32 hash, address token) public {
        require(_whitelistedTokens[token], "token not whitelisted");
        uint256 tokenAmount = this.estimateCreatePredictionCost(token);
        IBEP20(token).transferFrom(msg.sender, address(this), tokenAmount);

        address[] memory path = new address[](2);
        path[0] = token;
        path[1] = address(_fame);

        uint256 deadline = block.timestamp + 1;
        IBEP20(token).approve(address(_swapRouter), tokenAmount);
        _swapRouter.swapExactTokensForTokens(tokenAmount, _fameCostPerMint, path, address(this), deadline);

        _createPrediction(hash);
    }

    /**
     * @notice Creates a prediction using FAME tokens.
     * @param hash A unique hash representing the prediction.
     */
    function createPrediction(bytes32 hash) public {
        require(_fame.transferFrom(msg.sender, address(this), _fameCostPerMint), "failed to transfer from");
        _createPrediction(hash);
    }

    function _createPrediction(bytes32 hash) private _requirePhase(Phase.MintingPhase)  {
        require(hash != 0x00000000000000000000000000000000, "cannot submit empty hash");
        require( _predictionExists[hash] == false, "prediction must be unique");
        uint256 tokId = _mint(msg.sender, string(abi.encodePacked(urlPrefix, bytes32ToString(hash))));

        uint256 rewardPoolInput = _fameCostPerMint * _priceToRewardPercentage / 100;
        uint256 consolidationInput = _fameCostPerMint * _currentConsoliation / 100;

        _predictionExists[hash] = true;
        _roundTotalPool += rewardPoolInput;
        _unspendableBalance += rewardPoolInput + consolidationInput;
        _winNFTs[tokId] = consolidationInput;
        emit NewPrediction(hash, tokId);
    }

    /**
     * @notice Burns an NFT and transfers the corresponding reward to the NFT owner.
     * @param nftId The NFT token ID.
     */
    function burn(uint256 nftId) public override {
        require(this.ownerOf(nftId) == msg.sender, "only owner can burn for reward");
        require(_lastMaxNFT > nftId || _currentPhase == Phase.RewardsIdlePhase, "wait to the end of competition to burn the nft");
        uint256 reward = _winNFTs[nftId];
        require(_fame.transfer(msg.sender, reward), "failed to transfer reward");
        _unspendableBalance -= reward;
        _burn(nftId);
    }

    /**
     * @notice Returns the burn reward for the given NFT ID.
     * @param nftId The NFT token ID.
     * @return The burn reward.
     */
    function getBurnReward(uint256 nftId) public view returns(uint256) {
        require(_lastMaxNFT > nftId || _currentPhase == Phase.RewardsIdlePhase, "burn reward for this nft is still unknown");
        return _winNFTs[nftId];
    }

    /**
     * @notice Initializes a new prediction round with the specified parameters.
     * @param tokenCost The cost per prediction in FAME tokens.
     * @param priceToRewardPercentage The percentage of the cost to be added to the rewards pool.
     * @param consoliationPercentage The percentage of the cost to be used for consolidation rewards.
     * @param bracketsRewardsPercentage An array of size of 3, containing the rewards percentages for each bracket of winning spots (first, second, third place).
     */
    function initRound(uint256 tokenCost, uint256 priceToRewardPercentage, uint256 consoliationPercentage, uint256[3] calldata bracketsRewardsPercentage) public _requirePhase(Phase.RewardsIdlePhase) onlyOwner {
        require(priceToRewardPercentage + consoliationPercentage <= 100, "percentages cannot go above 100");
        require(bracketsRewardsPercentage[0] + bracketsRewardsPercentage[1] + bracketsRewardsPercentage[2] == 100, "brackets percentages must sum to 100");
        _currentPhase = Phase.MintingPhase;
        _fameCostPerMint = tokenCost;
        _bracketsRewardsPercentage[0] = bracketsRewardsPercentage[0];
        _bracketsRewardsPercentage[1] = bracketsRewardsPercentage[1];
        _bracketsRewardsPercentage[2] = bracketsRewardsPercentage[2];
        _priceToRewardPercentage = priceToRewardPercentage;
        _currentConsoliation = consoliationPercentage;
        emit PhaseMintingStart();
    }

    /**
     * @notice Starts the prediction game.
     */
    function startGame() public _requirePhase(Phase.MintingPhase) onlyOwner {
        _currentPhase = Phase.GamePhase;
        emit PhaseGameStart();
    }

    mapping(uint256 => bool) private _winners;

    /**
     * @notice Registers the winners for the current prediction round.
     * @param winningNftIds A 2D array containing the winning NFT IDs for each out of 3 brackets.
     */
    function registerWinners(uint256[][3] calldata winningNftIds) public _requirePhase(Phase.GamePhase) onlyOwner {
        for (uint8 bracket = 0; bracket < 3; bracket++) {
            uint256 prize = _bracketsRewardsPercentage[bracket] * _roundTotalPool / 100;

            for (uint256 i = 0; i < winningNftIds[bracket].length; i++) {
                require(_lastMaxNFT <= winningNftIds[bracket][i], "this NFT does not belog into current competition");
                require(winningNftIds[bracket][i] < _tokenIdCounter.current(), "this nft was not minted");
                require(_winners[winningNftIds[bracket][i]] == false, "duplicated nft in input");
                _winners[winningNftIds[bracket][i]] = true;

                _winNFTs[winningNftIds[bracket][i]] += prize / winningNftIds[bracket].length;
            }
        }

        _roundTotalPool = 0;
        _lastMaxNFT = _tokenIdCounter.current();
        _currentPhase = Phase.RewardsIdlePhase;
        emit PhaseIdleStart();
    }

    function _mint(address to, string memory uri) private returns(uint256) {
        uint256 tokenId = _tokenIdCounter.current();
        _tokenIdCounter.increment();
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, uri);
        return tokenId;
    }

    /**
     * @notice Returns the current FAME token price for creating a prediction.
     * @return The FAME token price.
     */
    function getBepPrice() public view returns (uint256) {
        return _fameCostPerMint;
    }

    /**
     * @notice Returns the address of the FAME token.
     * @return The FAME token address.
     */
    function getBepToken() public view returns(address) {
        return address(_fame);
    }

    /**
    * @notice Checks if the given token is whitelisted as an auxiliary token for creating predictions with.
    * @param token The address of the token to check.
    * @return A boolean indicating whether the token is whitelisted or not.
    */
    function isWhitelistedAuxToken(address token) public view returns(bool) {
        return _whitelistedTokens[token];
    }

    /**
    * @notice Sets the whitelist state for an auxiliary token.
    * @dev Only the contract owner can call this function.
    * @param token The address of the token to be whitelisted or removed from the whitelist.
    * @param state A boolean representing the desired whitelist state (true for whitelisted, false for not whitelisted).
    */
    function setAuxTokenWhitelistState(address token, bool state) public onlyOwner {
        require(_whitelistedTokens[token] != state); // make estimator fail, to avoid useless transaction.
        _whitelistedTokens[token] = state;
        emit TokenWhitelistChange(token, state);
    }

    /**
    * @notice Estimates the cost of creating a prediction with given token.
    * @param token The address of the auxiliary token to estimate the cost for.
    * @return The estimated cost in the given token WEI.
    */
    function estimateCreatePredictionCost(address token) external view returns (uint256) {
        require(_whitelistedTokens[token], "token not whitelisted");
        address[] memory path = new address[](2);
        path[0] = token;
        path[1] = address(_fame);

        uint[] memory amounts = _swapRouter.getAmountsIn(_fameCostPerMint, path);
        return amounts[0];
    }

    /**
    * @notice Returns the maximum burnable NFT ID.
    * @return The ID of the last burnable NFT.
    */
    function burnableMaxId() public view returns(uint256) {
        return _lastMaxNFT;
    }

    function _beforeTokenTransfer(address from, address to, uint256 tokenId, uint256 batchSize)
        internal
        override(ERC721, ERC721Enumerable)
    {
        super._beforeTokenTransfer(from, to, tokenId, batchSize);
    }

    function _burn(uint256 tokenId) internal override(ERC721, ERC721URIStorage) {
        super._burn(tokenId);
    }

    function tokenURI(uint256 tokenId)
        public
        view
        override(ERC721, ERC721URIStorage)
        returns (string memory)
    {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721, ERC721Enumerable)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }

    function bytes32ToString(bytes32 _bytes32) private pure returns (string memory) {
        uint8 i = 0;
        bytes memory bytesArray = new bytes(64);
        for (i = 0; i < bytesArray.length; i++) {

            uint8 _f = uint8(_bytes32[i/2] & 0x0f);
            uint8 _l = uint8(_bytes32[i/2] >> 4);

            bytesArray[i] = toByte(_l);
            i = i + 1;
            bytesArray[i] = toByte(_f);
        }
        return string(bytesArray);
    }

    function toByte(uint8 _uint8) private pure returns(bytes1) {
        if(_uint8 < 10) {
            return bytes1(_uint8 + 48);
        } else {
            return bytes1(_uint8 + 87);
        }
    }

    modifier _requirePhase(Phase phase) {
        require(phase == _currentPhase, "this function is not available at current phase");
        _;
    }
}
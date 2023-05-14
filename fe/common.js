const gamePhaseIdle       = 0
const gamePhasePredicting = 1
const gamePhaseOngoing    = 2
const gamePhaseDone       = 3

let signer = null;
let provider = null;
let p2eContract = null;
let bep20Contract = null;
let address = null;

function openUrl(url) {
    window.location.href = baseUrl + url
}

async function initMM(){
    if (p2eContract != null) return true;

    if (window.ethereum == null) {
        alert("metamask not detected")
        return false;
    }

    try{
        const accnts = await ethereum.request({ method: "eth_requestAccounts" })
        address = accnts[0];
        await window.ethereum.request({
            method: 'wallet_switchEthereumChain',
            params: [{ chainId: chainId }], // chainId must be in hexadecimal numbers
        });
    } catch (e) {
        console.log("failed to init MM, but MM found & installed")
        return
    }

    provider = new ethers.BrowserProvider(window.ethereum)
    signer = await provider.getSigner();
    bep20Contract = new ethers.Contract(BEP20Addr, IBEP20Abi, signer);
    p2eContract = new ethers.Contract(P2EAddr, p2eAbi, signer);
    console.log("mm init ok")
    return true;
}

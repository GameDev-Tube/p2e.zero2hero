let burnables = [];
let burnablesFetched = false;

async function fetchNfts() {
    if(burnablesFetched) return;
    burnablesFetched = true;
    burnables = [];
    let totalNfts = await p2eContract.balanceOf(signer.address)
    if(totalNfts == 0) { burnablesFetched = false; burnables = []; return; }
    let cutoff = await p2eContract.burnableMaxId()
    for(let i = 0; i < totalNfts; i++) {
        const nftId = await p2eContract.tokenOfOwnerByIndex(signer.address, i)
        const reward = await p2eContract.getBurnReward(nftId)
        if (nftId >= cutoff) { burnablesFetched = false; burnables = []; return; }
        burnables.push({
            nftId:nftId,
            reward: reward,
        })
    }
}

function updateSelectFromNftList() {
    const sel = document.getElementById("nft-select")
    const children = sel.children
    for (const child of children) {
        sel.removeChild(child);
    }
    for (let i = 0; i < burnables.length; i++) {
        const token = burnables[i];
        const newOption = document.createElement('option');
        newOption.value = token.nftId;
        newOption.textContent = "#"+token.nftId + " (for " + ethers.formatEther(token.reward) + ")"
        sel.appendChild(newOption);
    }
}

function showMessage(lbl) {
    document.getElementById("message-banner").style.display = ""
    document.getElementById("message-banner-lbl").innerText = lbl
}

function hideMessage() {
    document.getElementById("message-banner").style.display = "none"
    document.getElementById("prediction-container").style.display = ""
}

async function doBurn(e) {
    e.preventDefault()
    const sel = document.getElementById("nft-select");
    const idx = sel.value;
    console.log(idx)
    await(await p2eContract.burn(idx)).wait()
}

async function connectMMBtn() {
    if(!await initMM()) return;
    await fetchNfts();
    updateSelectFromNftList();
    document.getElementById("connect-container").style.display = "none"
    document.getElementById("burn-form").style.display = ""
}

window.addEventListener("load", function () {
    document.getElementById("connect-btn").addEventListener("click", connectMMBtn);
    document.getElementById("burn-btn").addEventListener("click", doBurn)
})
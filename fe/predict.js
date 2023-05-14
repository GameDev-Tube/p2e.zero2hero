let tokenWhitelist = [];

async function connectMMBtn() {
    if(!await initMM()) return;
    await fetchWhitelist();
    updateSelectFromWhitelist();
    document.getElementById("connect-container").style.display = "none"
    document.getElementById("form-container").style.display = ""
}

async function fetchWhitelist() {
    const whitelist = await (await fetch(baseUrl + "api/swap-tokens")).json()
    let i = 0
    for (const whitelistedAddress of whitelist.whitelist) {
        const contract = new ethers.Contract(whitelistedAddress, IBEP20Abi, signer);
        tokenWhitelist.push({
            address: whitelistedAddress,
            instance: contract,
            name: await contract.name(),
            symbol: await contract.symbol(),
            index: i
        })
        i++
    }
}

function updateSelectFromWhitelist() {
    const sel = document.getElementById("pancake-token-address")
    for (let i = 0; i < tokenWhitelist.length; i++) {
        const token = tokenWhitelist[i];
        const newOption = document.createElement('option');
        newOption.value = token.index;
        newOption.textContent = "[" + token.symbol + "] " + token.name +  " ("+ token.address +")";
        sel.appendChild(newOption);
    }
}

function getSelectedPancakeToken() {
    const sel = document.getElementById('pancake-token-address');
    const idx = sel.value;
    for(const whitelisted of tokenWhitelist) {
        if (whitelisted.index == idx) {
            return whitelisted
        }
    }
    throw "logic error"
}

function parseNum(str) {
    if(/^\d+$/.test(str) !== true)
        return false;
    const val = parseInt(str, 10);
    if (isFinite(val) && !isNaN(val)) {
        return val;
    }
    return false;
}

async function getPredHash() {
    let predA = parseNum(document.getElementById("predictionA").value)
    let predB = parseNum(document.getElementById("predictionB").value)
    let predC = parseNum(document.getElementById("predictionC").value)
    let predD = parseNum(document.getElementById("predictionD").value)
    let predE = parseNum(document.getElementById("predictionE").value)
    let predF = parseNum(document.getElementById("predictionF").value)

    if(predA === false || predB === false || predC === false || predD === false || predE === false || predF === false) {
        alert("empty fields or invalid numbers")
        return;
    }

    return await (await fetch(baseUrl + "api/predict", {
            body: JSON.stringify({
                choice_a: predA,
                choice_b: predB,
                choice_c: predC,
                choice_d: predD,
                choice_e: predE,
                choice_f: predF,
                game_edition: edition
            }),
            method: "POST",
            headers: { "Content-Type": "application/json" },
        })).json()
}

async function createPredictionPancake() {
    if(!await initMM()) return;

    try {
        showMessage("Preparing")
        const resp = await getPredHash();
        const auxContractData = getSelectedPancakeToken()
        const auxContract = auxContractData.instance
        let mintPrice = await p2eContract.estimateCreatePredictionCost(auxContractData.address)
        const allowance = await auxContract.allowance(signer.address, P2EAddr)

        // slippage
        mintPrice = (mintPrice * BigInt(103)) / BigInt(100)

        if(allowance < mintPrice) {
            document.getElementById("message-sub-banner-lbl").innerText = "if metamask asks for amount, click \"use default\" or input manually " + ethers.formatEther(mintPrice)
            const approveTx = await auxContract.approve(P2EAddr, mintPrice)
            document.getElementById("message-sub-banner-lbl").innerText = "waiting for approval to get confirmed in block..."
            await approveTx.wait()
        }

        document.getElementById("message-sub-banner-lbl").innerText = "waiting for mint transaction approval (metamask popup)..."
        const mintTx = await p2eContract.createPredictionAuxToken(resp.hash, auxContractData.address)
        document.getElementById("message-sub-banner-lbl").innerText = "waiting for mint to get confirmed in block..."
        await mintTx.wait()
        hideMessage()

    } catch (e) {
        alert(e)
        showMessage()
    }
}

async function createPrediction() {
    if(!await initMM()) return;

    try{
        showMessage("Preparing")
        const resp = await getPredHash();
        showMessage("Transacting, please consult MetaMask popups and wait for transactions...")
        const mintPrice = await p2eContract.getBepPrice()
        const allowance = await bep20Contract.allowance(signer.address, P2EAddr)
        if(allowance < mintPrice) {
            document.getElementById("message-sub-banner-lbl").innerText = "if metamask asks for amount, click \"use default\" or input manually " + ethers.formatEther(mintPrice)
            const approveTx = await bep20Contract.approve(P2EAddr, mintPrice)
            document.getElementById("message-sub-banner-lbl").innerText = "waiting for approval to get confirmed in block..."
            await approveTx.wait()
        }
        document.getElementById("message-sub-banner-lbl").innerText = "waiting for mint transaction approval (metamask popup)..."
        const mintTx = await p2eContract.createPrediction(resp.hash)
        document.getElementById("message-sub-banner-lbl").innerText = "waiting for mint to get confirmed in block..."
        await mintTx.wait()
        hideMessage()

    } catch (e) {
        console.log(e)
        alert(e)
        showMessage()
    }
}

function showMessage(lbl) {
    document.getElementById("message-banner").style.display = ""
    document.getElementById("message-banner-lbl").innerText = lbl
    document.getElementById("prediction-container").style.display = "none"
}

function hideMessage() {
    document.getElementById("message-banner").style.display = "none"
    document.getElementById("prediction-container").style.display = ""
    document.getElementById("message-sub-banner-lbl").innerText = ""
}

window.addEventListener("load", function () {
    if (phase !== gamePhasePredicting) {
        showMessage("Sorry, there is no ongoing prediction phase right now. Check back later!")
        return
    }

    document.getElementById("connect-btn").addEventListener("click", function () {
        connectMMBtn();
    })
    document.getElementById("predict-btn").addEventListener("click", function (ev){
        ev.preventDefault()
        createPrediction()
    })
    document.getElementById("predict-btn-pancake").addEventListener("click", function (ev){
        ev.preventDefault()
        createPredictionPancake()
    })
    document.getElementById("rng-btn").addEventListener("click", function (ev) {
        ev.preventDefault()
        document.getElementById("predictionA").value = Math.ceil(Math.random() * 1000)
        document.getElementById("predictionB").value = Math.ceil(Math.random() * 1000)
        document.getElementById("predictionC").value = Math.ceil(Math.random() * 1000)
        document.getElementById("predictionD").value = Math.ceil(Math.random() * 1000)
        document.getElementById("predictionE").value = Math.ceil(Math.random() * 1000)
        document.getElementById("predictionF").value = Math.ceil(Math.random() * 1000)
    })
})

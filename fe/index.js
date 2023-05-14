function showErrLbl(lbl) {
    document.getElementById("error-banner").style.display = ""
    document.getElementById("error-banner-lbl").innerText = lbl
    document.getElementById("top-3-board").style.display = "none"
}

function hideErrLbl() {
    document.getElementById("error-banner").style.display = "none"
    document.getElementById("top-3-board").style.display = ""
}
function previousGameBanner(){
    document.getElementById("previous-season-banner").style.display = ""
}

window.addEventListener("load", async function () {
    if ((phase !== gamePhaseOngoing && phase !== gamePhaseDone) && edition === 0) {
        showErrLbl("Not a single game was played out, yet, leaderboards are unavailable.")
        return
    }

    if(phase !== gamePhaseOngoing)
        previousGameBanner()

    const url = baseUrl + "api/leaderboards/3/0"
    try {
        const response = await (await fetch(url)).json()
        for (let i = 0; i < response.length; i++) {
            const ilbl = i + 1
            document.getElementById("board-" + ilbl).style.display = ""
            document.getElementById("top-" + ilbl + "-pred-a").innerText = response[i].choice_a
            document.getElementById("top-" + ilbl + "-pred-b").innerText = response[i].choice_b
            document.getElementById("top-" + ilbl + "-pred-c").innerText = response[i].choice_c
            document.getElementById("top-" + ilbl + "-pred-d").innerText = response[i].choice_d
            document.getElementById("top-" + ilbl + "-pred-e").innerText = response[i].choice_e
            document.getElementById("top-" + ilbl + "-pred-f").innerText = response[i].choice_f
            document.getElementById("top-" + ilbl + "-nft").innerText = "nft #" + response[i].nft_id
            document.getElementById("score-" + ilbl).innerText = response[i].score
        }
    } catch (e) {
        console.log(e)
        showErrLbl("failed to download leaderboards")
    }

})
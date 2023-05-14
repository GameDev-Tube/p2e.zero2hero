let currentPage = 1;
const perPage = 18;
let maxPage = null;

function showErrLbl(err){
    document.getElementById("error-banner-lbl").innerText = err
    document.getElementById("error-banner").style.display = ""
    document.getElementById("leaderboard-container").style.display = "none"
}
function hideErrLbl(){
    document.getElementById("error-banner").style.display = "none"
    document.getElementById("leaderboard-container").style.display = ""
}

function hideBtns() {
    document.getElementById("prev-pg-btn").style.display = "none"
    document.getElementById("next-pg-btn").style.display = "none"
}
function showBtns() {
    if(currentPage === 1) {
        const btnPrev = document.getElementById("prev-pg-btn").style.display = "none"
    } else {
        const btnPrev = document.getElementById("prev-pg-btn").style.display = ""
    }
    if(maxPage === null || currentPage < maxPage) {
        const btnNext = document.getElementById("next-pg-btn").style.display = ""
    } else {
        const btnNext = document.getElementById("prev-pg-btn").style.display = ""
    }
}

function previousGameBanner(){
    document.getElementById("previous-season-banner").style.display = ""
}

async function fetchLeaderboard() {
    hideBtns()
    if ((phase !== gamePhaseOngoing && phase !== gamePhaseDone) && edition === 0) {
        showBtns()
        showErrLbl("Not a single game was played out, yet, leaderboards are unavailable.")
        return
    }
    if(phase !== gamePhaseOngoing)
        previousGameBanner()

    const limit = perPage;
    const offset = (currentPage-1) * perPage;

    const els = getLeaderboardElements()
    const url = baseUrl + "api/leaderboards/"+limit+"/"+offset
    try {
        const response = await (await fetch(url)).json()
        // we went out of boundaries
        if(response.length < perPage) {
            maxPage = currentPage
            // special case if len == 0, means previous page was last, and just so happened to be full
            if(response.length === 0) {
                maxPage--;
                currentPage--
                return
            }
        }
        for (let i = 0; i < perPage; i++) {
            if(response[i]) {
                els[i].position.innerText = response[i].position
                els[i].eoa.innerText = response[i].owner
                els[i].score.innerText = response[i].score
                els[i].container.style.display = ""
            } else {
                els[i].container.style.display = "none"
            }
        }
    } catch (e) {
        console.log(e)
        showErrLbl("failed to download leaderboards")
        return
    }
    showBtns()
    hideErrLbl()
}

function getLeaderboardElements() {
    let output = [];
    const cont = document.getElementById("leaderboard")
    for(let i = 0; i < perPage; i++) {
        output.push({
            container: cont.children[i],
            position: cont.children[i].children[0],
            eoa: cont.children[i].children[1],
            score: cont.children[i].children[2],
        })
    }
    return output
}

window.addEventListener("load", async function () {
    const el = getLeaderboardElements()
    for (const elem of el) {
        elem.container.style.display = "none"
    }
    fetchLeaderboard()
    document.getElementById("prev-pg-btn").style.display = "none"
    document.getElementById("prev-pg-btn").addEventListener("click", function () {
        if(currentPage === 1) return
        currentPage--
        if(currentPage === 1) document.getElementById("prev-pg-btn").style.display = "none"
        fetchLeaderboard()
    })
    document.getElementById("next-pg-btn").addEventListener("click", function () {
        if(maxPage === null || currentPage < maxPage) {
            currentPage++
            fetchLeaderboard()
        }
    })
})


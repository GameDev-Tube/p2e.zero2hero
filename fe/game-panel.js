let auth = false;

window.addEventListener("load", function () {
    for (const k of ["a", "b", "c", "d", "e", "f"]) {
        const min = document.getElementById("min-"+k)
        const max = document.getElementById("max-"+k)
        const swpr = document.getElementById("swiper"+(k.toUpperCase()))
        const handle = function () {
            swpr.min = min.value
            swpr.max = max.value
        }

        min.addEventListener("change", handle)
        max.addEventListener("change", handle)
    }

    let ws = null;
    document.getElementById("login-btn").addEventListener("click", function (e) {
        e.preventDefault()

        const authRespHandle = function (e){
            const res = JSON.parse(e.data)
            ws.removeEventListener("message", authRespHandle)
            if (res.event === "ok") {
                auth = true;

                res.data.a = parseInt(res.data.a, 10)
                res.data.b = parseInt(res.data.b, 10)
                res.data.c = parseInt(res.data.c, 10)
                res.data.d = parseInt(res.data.d, 10)
                res.data.e = parseInt(res.data.e, 10)
                res.data.f = parseInt(res.data.f, 10)

                document.getElementById("val-a").value = res.data.a
                document.getElementById("swiperA").value = res.data.a

                document.getElementById("val-b").value = res.data.b
                document.getElementById("swiperB").value = res.data.b

                document.getElementById("val-c").value = res.data.c
                document.getElementById("swiperC").value = res.data.c

                document.getElementById("val-d").value = res.data.d
                document.getElementById("swiperD").value = res.data.d

                document.getElementById("val-e").value = res.data.e
                document.getElementById("swiperE").value = res.data.e

                document.getElementById("val-f").value = res.data.f
                document.getElementById("swiperF").value = res.data.f

                document.getElementById("login-container").style.display = "none"
                document.getElementById("panel-container").style.display = ""

                const changeCallbackSwiper = function () {
                    const update = {
                        event:"update-game-outcomes",
                        data: {
                            a:document.getElementById("swiperA").value,
                            b:document.getElementById("swiperB").value,
                            c:document.getElementById("swiperC").value,
                            d:document.getElementById("swiperD").value,
                            e:document.getElementById("swiperE").value,
                            f:document.getElementById("swiperF").value
                        }
                    }
                    ws.send(JSON.stringify(update))
                }
                const changeCallbackDirect = function () {
                    const update = {
                        event:"update-game-outcomes",
                        data: {
                            a:document.getElementById("val-a").value,
                            b:document.getElementById("val-b").value,
                            c:document.getElementById("val-c").value,
                            d:document.getElementById("val-d").value,
                            e:document.getElementById("val-e").value,
                            f:document.getElementById("val-f").value
                        }
                    }
                    ws.send(JSON.stringify(update))
                }

                document.getElementById("swiperA").addEventListener("change", changeCallbackSwiper)
                document.getElementById("swiperB").addEventListener("change", changeCallbackSwiper)
                document.getElementById("swiperC").addEventListener("change", changeCallbackSwiper)
                document.getElementById("swiperD").addEventListener("change", changeCallbackSwiper)
                document.getElementById("swiperE").addEventListener("change", changeCallbackSwiper)
                document.getElementById("swiperF").addEventListener("change", changeCallbackSwiper)

                document.getElementById("val-a").addEventListener("change", changeCallbackDirect)
                document.getElementById("val-b").addEventListener("change", changeCallbackDirect)
                document.getElementById("val-c").addEventListener("change", changeCallbackDirect)
                document.getElementById("val-d").addEventListener("change", changeCallbackDirect)
                document.getElementById("val-e").addEventListener("change", changeCallbackDirect)
                document.getElementById("val-f").addEventListener("change", changeCallbackDirect)

                ws.addEventListener("close", function (e) { console.log(e)
                    document.getElementById("login-container").style.display = ""
                    document.getElementById("panel-container").style.display = "none"
                    document.getElementById("swiperA").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("swiperB").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("swiperC").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("swiperD").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("swiperE").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("swiperF").removeEventListener("change", changeCallbackSwiper)
                    document.getElementById("val-a").removeEventListener("change", changeCallbackDirect)
                    document.getElementById("val-b").removeEventListener("change", changeCallbackDirect)
                    document.getElementById("val-c").removeEventListener("change", changeCallbackDirect)
                    document.getElementById("val-d").removeEventListener("change", changeCallbackDirect)
                    document.getElementById("val-e").removeEventListener("change", changeCallbackDirect)
                    document.getElementById("val-f").removeEventListener("change", changeCallbackDirect)
                    auth = false
                })
            } else {
                if (res.data !== undefined && res.data !== null && res.data.msg !== undefined && res.data.msg !== null ) {
                    alert(res.data.msg);
                } else {
                    alert("comms failed")
                }
            }
        }

        if(ws == null){
            ws = new WebSocket(wsUrl);
            ws.addEventListener("open", function () {
                // auth
                ws.send(JSON.stringify({key: document.getElementById("login-key").value}))
                ws.addEventListener("message", authRespHandle)
            })
        }
    })

    // bind swipers with value fields
    const swiperValCb = function (predId, updateSwiper) {
        return function () {
            const field = document.getElementById("val-"+predId.toLowerCase())
            const swiper = document.getElementById("swiper"+predId.toUpperCase())
            if (updateSwiper) {
                swiper.value = field.value
            } else {
                field.value = swiper.value
            }
        }
    }
    document.getElementById("swiperA").addEventListener("change", swiperValCb("A", false))
    document.getElementById("swiperB").addEventListener("change", swiperValCb("B", false))
    document.getElementById("swiperC").addEventListener("change", swiperValCb("C", false))
    document.getElementById("swiperD").addEventListener("change", swiperValCb("D", false))
    document.getElementById("swiperE").addEventListener("change", swiperValCb("E", false))
    document.getElementById("swiperF").addEventListener("change", swiperValCb("F", false))
    document.getElementById("val-a").addEventListener("change", swiperValCb("A", true))
    document.getElementById("val-b").addEventListener("change", swiperValCb("B", true))
    document.getElementById("val-c").addEventListener("change", swiperValCb("C", true))
    document.getElementById("val-d").addEventListener("change", swiperValCb("D", true))
    document.getElementById("val-e").addEventListener("change", swiperValCb("E", true))
    document.getElementById("val-f").addEventListener("change", swiperValCb("F", true))


    document.getElementById("phase-btn").addEventListener("click", async function (e) {
        e.preventDefault()
        if(!await initMM()) return
        document.getElementById("phase-form-container").style.display = ""
        document.getElementById("panel-container").style.display = "none"
    })
    document.getElementById("finish-btn").addEventListener("click", async function(e) {
        e.preventDefault()
        const spots = {
            f: parseInt(document.getElementById("spots-1").value),
            s: parseInt(document.getElementById("spots-2").value),
            t: parseInt(document.getElementById("spots-3").value)
        }
        const totalSpots = spots.f + spots.s + spots.t
        const url = baseUrl + "api/leaderboards/"+totalSpots+"/0"
        try {
            const response = await (await fetch(url)).json()
            let prms = [[],[],[]];
            for (let i = 0; i < response.length; i++) {
                if(i < spots.f) {
                    prms[0].push(response[i].nft_id)
                } else if (i < spots.f + spots.s) {
                    prms[1].push(response[i].nft_id)
                } else {
                    prms[2].push(response[i].nft_id)
                }
            }

            let tx = await p2eContract.registerWinners(prms)
            await tx.wait()

        } catch (e) {
            alert("failed")
        }
    })
});
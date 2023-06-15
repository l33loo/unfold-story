document.addEventListener("DOMContentLoaded", () => {
    const url = window.location.href
    const urlParts = url.split("/")
    const uuid = urlParts[urlParts.length - 1]
    const playerName = localStorage.getItem("name")
    if (playerName !== null && playerName.length > 0) {
        const nameInput = document.getElementById("name")
        nameInput.value = playerName
        console.log(`name!! localStorage: ${playerName}, inputVal: ${nameInput.value}`)
    }

    const body = document.getElementsByTagName("body")[0]

    const joinEventHandler = (e) => {
        e.preventDefault()

        const nameInput = document.getElementById("name")
        const name = nameInput.value
        localStorage.setItem("name", name)

        // We can’t use XMLHttpRequest or fetch to make this kind of HTTP-request,
        // because JavaScript is not allowed to set these headers.
        // https://javascript.info/websocket
        const ws = new WebSocket(`ws://localhost:8080/ws/${uuid}`)
        ws.onerror = (e) => {
            console.log("websocket error:")
            console.dir(e)
        }

        ws.onclose = (e) => {
            console.log(`websocket closed with code ${e.code}`)
        }

        ws.onopen = () => {
            console.log("websocket open <3")
            const msg = {NewPlayer: name}
            ws.send(JSON.stringify(msg))
            const user = document.createElement("p")
            user.textContent = `You are ${name}`
            const players = document.getElementById("players")
            body.insertBefore(user, players)
            const joinForm = document.getElementById("join")
            const game = document.createElement("div")
            game.setAttribute("id", "game")
            const gameLines = document.createElement("ol")
            gameLines.setAttribute("id", "game-lines")
            // const gameForm = document.createElement("form")
            // gameForm.setAttribute("id", "game-form")
            // const gameInput = document.createElement("input")
            // gameInput.setAttribute("id", "text")
            // const gameButton = document.createElement("button", {id: "submit-line"})
            // gameButton.textContent = "Submit a line"
            // gameForm.appendChild(gameInput)
            // gameForm.appendChild(gameButton)
            game.appendChild(gameLines)
            // game.appendChild(gameForm)
            joinForm.replaceWith(game)
        }

        ws.onmessage = (e) => {
            const gameLines = document.getElementById("game-lines")
            console.log("ws data <3:")
            // TODO: Handle PINGs and PONGs
            if (e.data === "PING" || e.data === "PONG") {
                return
            }
            
            let msgObj 
            try {
                msgObj = JSON.parse(e.data)
            } catch(e) {
                console.log("JSON ERROR <3 ", e)
            }
            console.dir(msgObj)
            for (const property in msgObj) {
                const val = msgObj[property]
                switch (property) {
                    case "Players":
                        console.log("PROPERTY <3: ", property)
                        const playerList = document.getElementById("player-list")
                        playerList.innerHTML = ''

                        val.forEach(p => {
                            const playerItem = document.createElement("li")
                            playerItem.textContent = p
                            playerList.appendChild(playerItem)
                        })

                        // First or only user
                        console.log("VAL <3: ", val)
                        if (val.length === 1) {
                            const gameForm = document.createElement("form")
                            gameForm.setAttribute("id", "game-form")
                            const gameInput = document.createElement("input")
                            gameInput.setAttribute("id", "text")
                            const gameButton = document.createElement("button")
                            gameButton.setAttribute("id", "submit-line")
                            gameButton.textContent = "Submit a line"
                            gameForm.appendChild(gameInput)
                            gameForm.appendChild(gameButton)
                            const game = document.getElementById("game")
                            game.appendChild(gameForm)
                            // const gameLines = document.getElementById("game-lines")
                            // const line = document.createElement("li")
                            // line.textContent = val
                            // gameLines.appendChild(line)
                            const submitLineEventHandler = ev => {
                                ev.preventDefault()
                                const userObj = {
                                    Broadcast: {
                                        LineAuthor: name
                                    }
                                }
                                ws.send(JSON.stringify(userObj))

                                const line = gameInput.value
                                const lineObj = {
                                    NextPlayer: line
                                }
                                ws.send(JSON.stringify(lineObj))

                                console.log("Remove!")
                                console.dir(gameForm)
                                console.dir(ev.currentTarget)
                                ev.currentTarget.remove()
                            }
                            gameForm.addEventListener("submit", submitLineEventHandler)
                        }
                        break
                    case "LineAuthors":
                        for (i = 0; i < val.length; i++) {
                            const line = document.createElement("li")
                            line.textContent = `Line by ${val[i]}`
                            gameLines.appendChild(line)
                        }
                        break
                    case "Forward":
                        console.log("PROPERTY <3: ", property)
                        // const gameForm = document.createElement("form")
                        // gameForm.setAttribute("id", "game-form")
                        // const gameInput = document.createElement("input")
                        // gameInput.setAttribute("id", "text")
                        // const gameButton = document.createElement("button")
                        // gameButton.setAttribute("id", "submit-line")
                        // gameButton.textContent = "Submit a line"
                        // gameForm.appendChild(gameInput)
                        // gameForm.appendChild(gameButton)
                        // game.appendChild(gameForm)
                        if (!!val.LineAuthor) {
                            const line = document.createElement("li")
                            line.textContent = `Line by ${val.LineAuthor}`
                            gameLines.appendChild(line)
                        }

                        if (!!val.NextPlayer) {
                            const lastLine = gameLines.lastChild
                            const line = document.createTextNode(`: ${val.NextPlayer}`)
                            lastLine.appendChild(line)
                            // const line = document.createElement("li")
                            // line.textContent = val.NextPlayer
                            // gameLines.appendChild(line)
                            const gameForm = document.createElement("form")
                            gameForm.setAttribute("id", "game-form")
                            const gameInput = document.createElement("input")
                            gameInput.setAttribute("id", "text")
                            const gameButton = document.createElement("button")
                            gameButton.setAttribute("id", "submit-line")
                            gameButton.textContent = "Submit a line"
                            gameForm.appendChild(gameInput)
                            gameForm.appendChild(gameButton)
                            const game = document.getElementById("game")
                            game.appendChild(gameForm)
                            const submitLineEventHandler = ev => {
                                ev.preventDefault()

                                const lineAuthor = {
                                    Broadcast: {
                                        LineAuthor: name
                                    }
                                }
                                ws.send(JSON.stringify(lineAuthor))

                                const line = gameInput.value
                                const lineObj = {
                                    NextPlayer: line
                                }
                                ws.send(JSON.stringify(lineObj))

                                console.log("Remove!")
                                console.dir(gameForm)
                                ev.currentTarget.remove()
                            }
                            gameForm.addEventListener("submit", submitLineEventHandler)
                        }
                        break
                    case "TheEnd":
                        const gameForm = document.getElementById("game-form")
                        if (gameForm) {
                            gameForm.remove()
                        }
                        gameLines.innerHTML = ""
                        val.forEach(l => {
                            const li = document.createElement("li")
                            li.textContent = `Line by ${l.Author}: ${l.Text}`
                            gameLines.appendChild(li)
                        })
                        const theEnd = document.createElement("div")
                        theEnd.textContent = "The End!"
                        const game = document.getElementById("game")
                        game.appendChild(theEnd)
                        break
                    // case "Start":
                    //     const startForm = document.createElement("form")
                    //     const startButton = document.createElement("button")
                    //     startButton.setAttribute("id", "start")
                    //     startButton.setAttribute("type", "submit")
                    //     startButton.setAttribute("value", "Start game")
                    //     startButton.textContent = "Start game"
                    //     // const hiddenInput = document.createAttribute("input")
                    //     // hiddenInput.setAttribute("type", "hidden")
                    //     startForm.appendChild(startButton)
                    //     // startForm.appendChild(hiddenInput)
                    //     body.appendChild(startForm)

                    //     startForm.addEventListener("submit", e => {
                    //         e.preventDefault()
                    //         startForm.remove()
                    //     })
                }
            }
        }
    }

    const join = document.getElementById("join")
    join.addEventListener("submit", joinEventHandler)
})
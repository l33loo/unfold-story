document.addEventListener("DOMContentLoaded", () => {
    const playerName = localStorage.getItem("name")
    if (playerName !== null && playerName.length > 0) {
        const nameInput = document.getElementById("name")
        nameInput.value = playerName
        console.log(`name!! localStorage: ${playerName}, inputVal: ${nameInput.value}`)
    }

    const joinEventHandler = (e) => {
        e.preventDefault()
        const nameInput = document.getElementById("name")
        const name = nameInput.value
        localStorage.setItem("name", name)

        const playerList = document.getElementsByTagName("ul")[0]

        // We can’t use XMLHttpRequest or fetch to make this kind of HTTP-request,
        // because JavaScript is not allowed to set these headers.
        // https://javascript.info/websocket
        let ws = new WebSocket("ws://localhost:8080/ws")
        ws.onerror = (e) => {
            console.log("websocket error:")
            console.dir(e)
        }

        ws.onclose = (e) => {
            console.log(`websocket closed with code ${e.code}`)
        }

        ws.onopen = () => {
            console.log("websocket open <3")
            ws.send(`${name}`)
        }

        ws.onmessage = (e) => {
            console.log("ws data <3:")
            console.dir(e.data)

            const playerItem = document.createElement("li")
            playerItem.textContent = e.data
            playerList.appendChild(playerItem)
        }
    }

    const join = document.getElementById("join")
    join.addEventListener("submit", joinEventHandler)
})
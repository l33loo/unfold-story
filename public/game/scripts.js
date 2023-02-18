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
            ws.send(name)
        }
    }

    const join = document.getElementById("join")
    join.addEventListener("submit", joinEventHandler)
})
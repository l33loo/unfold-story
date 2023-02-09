document.addEventListener("DOMContentLoaded", () => {
    fetch("/ws", {
        method: "GET",
        headers: {
            "Host": "localhost",
            "Upgrade": "websocket",
            "Connection": "Upgrade",
            "Sec-WebSocket-Key": "hello",
            "Origin": "localhost",
            "Sec-WebSocket-Protocol": "chat, superchat",
            "Sec-WebSocket-Version0": "13"
        }
    }).then(() => {
        console.log("fetch then <3")

        console.log("window load <3")
    const playerName = localStorage.getItem("name")
    if (playerName !== null && playerName.length > 0) {
        const nameInput = document.getElementById("name")
        nameInput.value = playerName
        console.log(`name!! localStorage: ${playerName}, inputVal: ${nameInput.value}`)
    }

    const joinEventHandler = () => {
        const nameInput = document.getElementById("name")
        localStorage.setItem("name", nameInput.value)
    }

    const join = document.getElementById("join")
    join.addEventListener("submit", joinEventHandler)
    })
})
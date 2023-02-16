document.addEventListener("DOMContentLoaded", () => {
    fetch("/ws", {
        method: "GET",
        headers: {
            // As per RFC6455:
            // The client includes the hostname in the |Host| header field of its
            // handshake as per [RFC2616], so that both the client and the server
            // can verify that they agree on which host is in use.
            "Host": "localhost:8080",
            "Upgrade": "websocket",
            "Connection": "Upgrade",
            "Sec-WebSocket-Key": "hello",
            "Origin": "localhost",
            "Sec-WebSocket-Protocol": "chat, superchat",
            "Sec-WebSocket-Version": "13"
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
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
    }).then((resp) => {
        // If the status code received from the server is not 101, the
        // client handles the response per HTTP [RFC2616] procedures.  In
        // particular, the client might perform authentication if it
        // receives a 401 status code; the server might redirect the client
        // using a 3xx status code (but clients are not required to follow
        // them), etc.  Otherwise, proceed as follows.
        const status = resp.status
        if (status !== 101) {
            throw new Error(`wrong HTTP status: ${status} instead of 101`)
        }

        // If the response lacks an |Upgrade| header field or the |Upgrade|
        // header field contains a value that is not an ASCII case-
        // insensitive match for the value "websocket", the client MUST
        // _Fail the WebSocket Connection_.
        const headers = resp.headers
        console.log("headers <3")
        console.dir(headers)
        const upgrade = headers.get("Upgrade")
        if (upgrade !== "websocket") {
            throw new Error(`error with Upgrade header: value ${upgrade} instead of 'websocket'`)
        }

        // If the response lacks a |Connection| header field or the
        // |Connection| header field doesn't contain a token that is an
        // ASCII case-insensitive match for the value "Upgrade", the client
        // MUST _Fail the WebSocket Connection_.
        const connection = headers.get("Connection")
        if (connection !== "Upgrade") {
            throw new Error(`error with Connection header: value ${connection} instead of "Upgrade`)
        }

        // If the response lacks a |Sec-WebSocket-Accept| header field or
        // the |Sec-WebSocket-Accept| contains a value other than the
        // base64-encoded SHA-1 of the concatenation of the |Sec-WebSocket-
        // Key| (as a string, not base64-decoded) with the string "258EAFA5-
        // E914-47DA-95CA-C5AB0DC85B11" but ignoring any leading and
        // trailing whitespace, the client MUST _Fail the WebSocket
        // Connection_.
        const secWebSocketAccept = headers.get("Sec-WebSocket-Accept")
        if (secWebSocketAccept !== "Kfh9QIsMVZcl6xEPYxPHzW8SZ8w="){
            throw new Error(`invalid Sec-WebSocket-Accept header ${secWebSocketAccept}`)
        }
        
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
    }).catch((err) => {
        // handle error
        console.log(err.message)
    })
})
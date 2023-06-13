document.addEventListener("DOMContentLoaded", () => {
    const uuidInput = document.getElementById("new-game_uuid")
    const uuid = self.crypto.randomUUID()
    uuidInput.value = uuid
    const newGameForm = document.getElementById("new-game")
    // TODO: get to update when UUID is changed
    newGameForm.setAttribute("action", `/game/${uuid}`)
})
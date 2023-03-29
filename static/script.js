const dialogs = {
  'seal': 'dialog-1',
  'snitch': 'dialog-2',
  'birb': 'dialog-3',
  'c3po': 'dialog-4',
  'kevin': 'dialog-5',
  'crabs': 'dialog-6',
  'mcguirk': 'dialog-7'
};

function hideDialog(character, dialog) {
  dialog.classList.remove("popin");
  dialog.classList.add("popout");

  character.classList.remove("popin");
  character.classList.add("popout");
  character.classList.add("hide");
}

function showDialog(content, character, dialog) {
    dialog.classList.remove("popout");
    dialog.classList.remove("hide");
    dialog.classList.add("popin");

    dialog.innerHTML = content

    character.classList.remove("popout");
    character.classList.remove("hide");
    character.classList.add("popin");
}

function callback(mutationList, observer) {
  mutationList.forEach(function(mutation) {
    if (mutation.type === 'attributes' && mutation.attributeName === 'class') {

      // We should be able to check what class changed
      // handle class change
    }
  })
}

function bringCharactersToLife() {
  const updates = document.getElementById('updates');
  const socket = new WebSocket('ws://' + location.host + '/ws');
    
    let lip_sync_wrapper = document.getElementById("lip_sync_wrapper");


    // When the 
    // video.play()

    video.onended = (event) => {
      lip_sync_wrapper.classList.add("hide");
      console.log(
        "Video stopped either because it has finished playing or no further data is available."
      );
    };

  socket.addEventListener('message', event => {

    // This should only parse certain events?????
    let contents = event.data.split(/[ ,]+/)
    let eventName = contents[0]
    let character = contents[1];


    let dialogText = dialogs[character];

    // if (dialogText === undefined) {
    //   dialogText = "dialog-1"
    // }
    // if (character === undefined) {
    //   character = "seal"
    // }
    // if (dialogText === undefined && character === undefined) {
    //   // console.log("WTF: " + eventName);
    //   // return
    // }

    console.log(character + " | " + dialogText)

    let charDiv = document.getElementById(character)
    let dialog = document.getElementById(dialogText);

    if (eventName === "done") {
        console.log("Hide")
        hideDialog(charDiv, dialog)
    } else if (eventName == "dialog") {

      // Split the string based on whitespace
      const wordsArray = event.data.split(/\s+/);

      // Remove the first two elements
      wordsArray.shift();
      wordsArray.shift();

      // Concatenate the remaining elements back into a single string
      const newContent = wordsArray.join(' ');
      showDialog(newContent, charDiv, dialog)
    } else if (eventName == "start_animation") {
        const video_wrapper = document.getElementById("lip_sync_wrapper");
        console.log("ANIMAITON TIME")
        video.load();
        console.log("POST LOAD")
        video_wrapper.classList.remove("hide");
        console.log("POST HIDE")

        // we Might need to wait here for a second to load
        // video.muted = false;
        video.play();
        console.log("POST PLAY")
    } else if (eventName == "end_animation") {
      // const video = document.querySelector("#lip");
      // We need a way of trigger ending animations
    }
  })
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

const video = document.getElementById("lip_sync");
const observer = new MutationObserver(callback)
const options = {
  attributes: true
}
observer.observe(video, options)

bringCharactersToLife()

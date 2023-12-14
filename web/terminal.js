const BELL_SOUND = "data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4LjMyLjEwNAAAAAAAAAAAAAAA//tQxAADB8AhSmxhIIEVCSiJrDCQBTcu3UrAIwUdkRgQbFAZC1CQEwTJ9mjRvBA4UOLD8nKVOWfh+UlK3z/177OXrfOdKl7pyn3Xf//WreyTRUoAWgBgkOAGbZHBgG1OF6zM82DWbZaUmMBptgQhGjsyYqc9ae9XFz280948NMBWInljyzsNRFLPWdnZGWrddDsjK1unuSrVN9jJsK8KuQtQCtMBjCEtImISdNKJOopIpBFpNSMbIHCSRpRR5iakjTiyzLhchUUBwCgyKiweBv/7UsQbg8isVNoMPMjAAAA0gAAABEVFGmgqK////9bP/6XCykxBTUUzLjEwMKqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq";

const bodyElem = document.getElementById("body");
const wrapperElem = document.getElementById("wrapper");
const terminalElem = document.getElementById("terminal");

const ws = new WebSocket(`wss://${location.host}/ws`);

let connId = null;

function debounce(func, timeout = 300) {
    let timer;
    return (...args) => {
        clearTimeout(timer);
        timer = setTimeout(() => { func.apply(this, args); }, timeout);
    };
}

toastr.options.positionClass = "toast-bottom-left";
toastr.options.progressBar = true;

ws.addEventListener("open", () => {
    const fit = new FitAddon.FitAddon();
    const terminal = new Terminal({
        allowProposedApi: true,
        allowTransparency: true,
        cursorStyle: "bar",
        cursorInactiveStyle: "bar",
        altClickMovesCursor: true,
        cursorBlink: true,
        fontWeight: "normal",
        fontFamily: "monospace",
        fontSize: 16,
        tabStopWidth: 4,
        logLevel: "error"
    });
    terminal.loadAddon(fit);
    terminal.loadAddon(new WebLinksAddon.WebLinksAddon());
    terminal.open(terminalElem);

    // Communication
    terminal.onData((data) => {
        const encodedData = new TextEncoder().encode(data);
        ws.send(new Uint8Array([1, encodedData.length, ...encodedData]));
    });
    ws.addEventListener("message", async (event) => {
        const encodedMessage = new TextEncoder().encode(event.data);
        if(encodedMessage.length < 1) return;
        const messageData = new TextDecoder().decode(encodedMessage.slice(1));
        switch(encodedMessage[0]) {
            case 1:
                terminal.write(messageData);
                break;
            case 2:
                connId = messageData;
                break;
            case 3:
                const theme = JSON.parse(messageData);
                bodyElem.style.backgroundColor = theme.xTerm.backgroundColor;
                if(theme.bgImg) {
                    wrapperElem.classList.add("acrylic");
                    bodyElem.style.backgroundImage = `url(${theme.bgImg})`;
                    terminal.options.background = "#00000000";
                } else {
                    wrapperElem.classList.remove("acrylic");
                    bodyElem.style.backgroundImage = "";
                }
                terminal.options.theme = theme.xTerm;
                if(theme.customFont) {
                    const font = await new FontFace(theme.font, theme.customFont).load();
                    document.fonts.add(font);
                }
                terminal.options.fontFamily = theme.font;
                document.documentElement.style.fontFamily = theme.font;
                console.log(`Loaded theme ${theme.id}`);
                break;
            case 4:
                toastr.success(messageData);
                break;
            case 5:
                window.open(`${messageData}?connectionId=${connId}`, "_blank")
                break;
        }
    });
    terminal.onResize((event) => {
        ws.send(new Uint8Array([2, event.cols >> 8, event.cols, event.rows >> 8, event.rows]));
    });

    // Bell Sound :)
    const BELL_AUDIO = new Audio(BELL_SOUND);
    terminal.onBell(() => BELL_AUDIO.play());

    // Handle Ctrl-C & Ctrl-V in a sane way
    terminal.attachCustomKeyEventHandler((event) => {
        if(!event.ctrlKey || (event.key != "c" && event.key != "v")) return true;
        if(event.type == "keydown" && event.key == "c") {
            if(terminal.getSelection().length <= 0) return true;
            navigator.clipboard.writeText(terminal.getSelection());
            terminal.clearSelection();
        }
        if(event.type == "keydown" && event.key == "v") {
            navigator.clipboard.readText().then((txt) => {
                const encodedText = new TextEncoder().encode(txt);
                ws.send(new Uint8Array([1, encodedText.length, ...encodedText]));
            });
        }
        event.preventDefault();
        return false;
    });

    const resize = debounce(() => fit.fit(), 100);
    const resizeOb = new ResizeObserver(() => resize());
    resizeOb.observe(terminalElem);
    fit.fit();

    terminal.focus();
});

ws.addEventListener("close", () => {
    toastr.error("Websocket closed unexpectedly", "Error", { timeOut: 0, extendedTimeOut: 0})
});

// File Upload
document.addEventListener("dragover", (e) => {
    e.preventDefault();
});

document.addEventListener("drop", (e) => {
    if(!connId) return;

    for(const file of e.dataTransfer.files) {
        let formData = new FormData();
        formData.append("file", file);
        fetch(`/upload?connectionId=${connId}`, {
            method: "POST",
            body: formData
        })
            .then(async (res) => {
                const msg = await res.text();
                if(res.status == 200) {
                    toastr.success(msg, "Success");
                } else {
                    toastr.error(msg, "Error");
                }
            })
            .catch(console.error);
    }
    e.preventDefault();
});
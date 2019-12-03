 //@tslint

console.log("script started");
itemMap = new Map();

window.addEventListener("load", ev => {
    console.log("opening websocket");
    let ws = new WebSocket("ws://"+window.location.host+"/websocket");
    ws.onopen = evt => {
        console.log("websocket opened");
    }

    ws.onmessage = evt => {
        let data = JSON.parse(evt.data);
        console.log(data);
        let update = itemMap.get(data.name);
        if (update)
            update(data);
    }

    ws.onclose = ect => {
        console.log("websocket closed");
        ws = null;
    }
});

function openMenu() {
    document.getElementById("sidemenu").style.width = "33%";
    document.getElementById("main").style.marginLeft = "33%";
}


function closeMenu() {
    document.getElementById("sidemenu").style.width = "0%";
    document.getElementById("main").style.marginLeft = "0%";
}

function invokeTemplates () {
    Array.prototype.slice.call(document.querySelectorAll('*[id]')).filter(node => node.id.startsWith("item-")).forEach(element => {
        console.log(element);
        let splitId = element.id.split("-", 3);
        let type = splitId[1];
        let itemName = splitId[2];
        let itmUpdate = undefined;
        switch (type) {
            case "text":
                {
                let txt = document.getElementsByTagName("template")[1];
                console.log(txt);
                let d = txt.content.querySelector("span");
                console.log(d);
                let node = document.importNode(d, true);
                element.appendChild(node);
                itmUpdate = item => {
                    console.log(node);
                    let nodeElements = node.getElementsByTagName("div");
                    console.log(nodeElements);
                    nodeElements[0].innerText = item.label + ": " + item.prefix + item.state + item.suffix
                }
                }
                break;
            
            case "switch":
                {
                let btn = document.getElementsByTagName("template")[0];
                console.log(btn);
                let d = btn.content.querySelector("span");
                let node = document.importNode(d, true);
                node.getElementsByTagName("button")[0].onclick = (evt) => setItemState(itemName, "ON");
                node.getElementsByTagName("button")[1].onclick = (evt) => setItemState(itemName, "OFF");
                element.appendChild(node);
                }
                break;
            case "rgb":
                {
                let rgb = document.getElementsByTagName("template")[2];
                console.log("rgb");
                let inp = rgb.content.querySelector("span");
                let node = document.importNode(inp, true);
                node.getElementsByTagName("input")[0].addEventListener("change", evt => {
                    let colorString = evt.target.value;
                    let rgbComponents = {red: colorString.slice(1,3), green: colorString.slice(3,5), blue: colorString.slice(5,7)};
                    let rgbValues = {red: parseInt("0x"+rgbComponents.red), green: parseInt("0x"+rgbComponents.green), blue: parseInt("0x"+rgbComponents.blue)}
                    let rgbString = rgbValues.red + "," + rgbValues.green + "," + rgbValues.blue;
                    console.log(colorString);
                    console.log(rgbValues);
                    console.log(rgbString);
                    setItemState(itemName, rgbString);
                }, false);
                element.appendChild(node);
                }
                break;
            default:
                console.log("unknown template type " + type);
                break;
        }
        itemMap.set(itemName, itmUpdate);
    });
    console.log(itemMap);
    getItems();
}

function getItems() {
    const Http = new XMLHttpRequest();
    const url="/rest/items";
    Http.open("GET", url);
    Http.send();
    Http.onreadystatechange=(e)=>{
    let text = Http.responseText;
    let items = JSON.parse(text);
    console.log(items);
    itemMap.forEach((update, name) => {
        console.log(name);
        console.log(update);
        let itm = items.filter(it => name === it.name)[0];
        if (update)
            update(itm);
    });
}
}

function setItemState(name, state) {
    const Http = new XMLHttpRequest();
    const url="/rest/items/"+name+"/state";
    Http.open("POST", url);
    Http.send(state);
    Http.onreadystatechange=(e)=>{
        let text = Http.responseText;
        let items = JSON.parse(text);
        console.log(items);
    }
}

function loadSite(site) {
    console.log("loading site", site);
    const Http = new XMLHttpRequest();
    const url="/sites/" + site + "/html";
    console.log(url);
    Http.open("GET", url);
    Http.send();
    Http.onreadystatechange=(e)=>{
        if (Http.readyState == 4 && Http.status == 200)
        {
            let text = Http.responseText;
            itemMap = new Map();
            document.getElementById("main").innerHTML = text;
            Array.prototype.slice.call(document.querySelectorAll('*[id]')).filter(node => node.id.startsWith("item-")).forEach(element => {
                let scripts = element.getElementsByTagName("script");
                for ( i = 0; ; i++) {
                    let script = scripts[i];
                    if (script == null)
                        break;
                    eval(script.innerText);
                }
            });
            document.getElementById("title").innerText = site;
            console.log("new site (" + site +") loaded");
        }
    }
}
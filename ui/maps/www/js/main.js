var map;

// "static/assets/main.wasm"
let wasmInstance;
const go = new Go();
WebAssembly.instantiateStreaming(
  fetch("static/assets/main.wasm"),
  go.importObject
).then((result) => {
  go.run(result.instance)
    .then(() => {
      console.log("Wasm app exited. Refreshing the page...");
      location.reload();
    })
    .catch((err) => {
      console.error("Error running Wasm app:", err);
    });
  wasmInstance = result.instance;
});

function newMap() {
  const tmpMap = new maplibregl.Map({
    container: "map", // container's id or the HTML element in which MapLibre GL JS will render the map
    style: 'static/assets/map_style.json', // style URL
    center: [16.62662018, 49.2125578], // starting position [lng, lat]
    zoom: 2, // starting zoom
    minZoom: 1.5,
  });


  tmpMap.once("load", function () {
    tmpMap.addControl(new maplibregl.NavigationControl(), "top-right");

    // Load custom svg icon
    markerIconSize = [22, 28]
    largeMarkerIconSize = [26, 32]

    // Generic marker icon
    let genericMarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    genericMarkerIcon.src = "static/assets/generic_marker.png";
    genericMarkerIcon.onload = () => tmpMap.addImage("genericMarkerIcon", genericMarkerIcon);

    // WSPR marker icon
    let wsprMarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    wsprMarkerIcon.src = "static/assets/wspr_marker.png";
    wsprMarkerIcon.onload = () => tmpMap.addImage("wsprMarkerIcon", wsprMarkerIcon);
    
    // FT8 marker icon
    let ft8MarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    ft8MarkerIcon.src = "static/assets/ft8_marker.png";
    ft8MarkerIcon.onload = () => tmpMap.addImage("ft8MarkerIcon", ft8MarkerIcon);

    let ft4MarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    ft4MarkerIcon.src = "static/assets/ft4_marker.png";
    ft4MarkerIcon.onload = () => tmpMap.addImage("ft4MarkerIcon", ft4MarkerIcon);

    // JT65 marker icon
    let jt65MarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    jt65MarkerIcon.src = "static/assets/jt65_marker.png";
    jt65MarkerIcon.onload = () => tmpMap.addImage("jt65MarkerIcon", jt65MarkerIcon);

    // UDARP marker icon
    let udarpMarkerIcon = new Image(markerIconSize[0], markerIconSize[1]);
    udarpMarkerIcon.src = "static/assets/udarp_marker.png";
    udarpMarkerIcon.onload = () => tmpMap.addImage("udarpMarkerIcon", udarpMarkerIcon);

    // Own marker icon
    let ownMarkerIcon = new Image(largeMarkerIconSize[0], largeMarkerIconSize[1]);
    ownMarkerIcon.src = "static/assets/own_marker.png";
    ownMarkerIcon.onload = () => tmpMap.addImage("ownMarkerIcon", ownMarkerIcon);
  

    map = tmpMap;
    return tmpMap;
  });
  return tmpMap;
}

// First load
map = newMap();

// Function that hides div with class .nav by using animation
function closeNav() {
  $(".nav").animate(
    {
      // width: 0,
      opacity: 0,
    },
    700
  );
}

// Function that shows div with class .nav by using animation
function openNav() {
  $(".nav").animate(
    {
      opacity: 1,
    },
    700
  );
}

// Function that listens on .hb menu button and toggles .nav
$(".hb").click(function () {
  // If .nav opacity is 0, then show it
  if ($(".nav").css("opacity") == 0) {
    document.getElementById("hamburgerButtonOpen").beginElement();
    openNav();
  }
  // If .nav width is not 0, then hide it
  else {
    document.getElementById("hamburgerButtonClose").beginElement();
    closeNav();
  }
});

function createLayer(layerName, markerIconName) {
  map.on("load", function () {
    map.addSource(layerName, {
      type: "geojson",
      data: null,
    });

    // Add a layer showing the places.
    map.addLayer({
      id: layerName,
      type: "symbol",
      source: layerName,
      layout: {
        "icon-image": markerIconName,
        "icon-overlap": "always",
      },
    });

    // Create a popup, but don't add it to the map yet.
    var popup = new maplibregl.Popup({
      closeButton: false,
      closeOnClick: false,
    });

    map.on("mouseenter", layerName, function (e) {
      // Change the cursor style as a UI indicator.
      map.getCanvas().style.cursor = "pointer";

      var coordinates = e.features[0].geometry.coordinates.slice();
      var description = e.features[0].properties.description;

      // Ensure that if the map is zoomed out such that multiple
      // copies of the feature are visible, the popup appears
      // over the copy being pointed to.
      while (Math.abs(e.lngLat.lng - coordinates[0]) > 180) {
        coordinates[0] += e.lngLat.lng > coordinates[0] ? 360 : -360;
      }

      // Populate the popup and set its coordinates
      // based on the feature found.
      popup.setLngLat(coordinates).setHTML(description).addTo(map);
    });

    map.on("mouseleave", layerName, function () {
      map.getCanvas().style.cursor = "";
      popup.remove();
    });
  });
}

// Function which undos everything that createLayer(layerName) does
function removeLayer(layerName) {
  // Check if layers and sources exist
  if (map.getLayer(layerName) != undefined) {
    map.off("mouseenter", layerName);
    map.off("mouseleave", layerName);
    map.removeLayer(layerName);
  }
  if (map.getSource(layerName) != undefined) {
    map.removeSource(layerName);
  }
}

// Start menu in open state
document.getElementById("hamburgerButtonOpen").beginElement();

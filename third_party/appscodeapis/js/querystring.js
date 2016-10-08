"use strict";
function decodeUriParam(value) {
    var values = value.split("=");
    return { "key": decodeURIComponent(values[0]), "value": decodeURIComponent(values[1]) };
}
function mapJSONToUriParams(data, encode, prefix) {
    if (encode === void 0) { encode = true; }
    if (prefix === void 0) { prefix = ""; }
    function JSONToUriParams(data, encode, prefix, call) {
        var map = [];
        if (data instanceof Array) {
            for (var ik = 0; ik < data.length; ik++) {
                map.push(JSONToUriParams(data[ik], encode, prefix + "[" + ik + "]", call + 1));
            }
        }
        else if (data instanceof Object) {
            for (var k in data) {
                var sep = "";
                //not empty
                if (prefix !== "") {
                    if (prefix.slice(-1) !== "]") {
                        sep = ":";
                    }
                }
                map.push(JSONToUriParams(data[k], encode, prefix + sep + k, call + 1));
            }
        }
        else {
            map.push(prefix + "=" + encodeURIComponent(data));
        }
        if (call == 0 && encode == true) {
            for (var i = 0; i < map.length; i++) {
                map[i] = encodeURIComponent(map[i]);
            }
        }
        return map.join("&");
    }
    return JSONToUriParams(data, encode, prefix, 0);
}
function mapObjectKey(key, value, object) {
    var indexOfObjectSep = key.indexOf(":");
    var indexOfArray = key.indexOf("[");
    if ((indexOfObjectSep > -1 && indexOfObjectSep < indexOfArray) || (indexOfObjectSep > -1 && indexOfArray === -1)) {
        var extractedKey = key.substr(0, indexOfObjectSep);
        var remainingKey = key.substr(indexOfObjectSep + 1);
        if (!(extractedKey in object)) {
            object[extractedKey] = {};
        }
        if (remainingKey === "") {
            object[extractedKey] = value;
        }
        else {
            return mapObjectKey(remainingKey, value, object[extractedKey]);
        }
    }
    else if ((indexOfArray > -1 && indexOfArray < indexOfObjectSep) || (indexOfArray > -1 && indexOfObjectSep === -1)) {
        var extractedKey = key.substr(0, indexOfArray);
        var remainingKey = key.substr(key.indexOf("]") + 1);
        if (!(extractedKey in object)) {
            object[extractedKey] = [];
        }
        var index = parseInt(key.substr(indexOfArray + 1, key.indexOf("]") - 1));
        if (!(index in object[extractedKey])) {
            object[extractedKey][index] = {};
        }
        if (remainingKey === "") {
            object[extractedKey][index] = value;
        }
        else {
            return mapObjectKey(remainingKey, value, object[extractedKey][index]);
        }
    }
    else {
        object[key] = value;
    }
}
function mapUriParamsToJSON(uriString, decodeTwice) {
    if (decodeTwice === void 0) { decodeTwice = true; }
    function UriParamsToJSON(uriString, decodeTwice, object, call) {
        if (call === 0) {
            if (decodeTwice) {
                uriString = decodeURI(uriString);
            }
            var decodedUriStringArray = decodeURIComponent(uriString).split("&");
            for (var key in decodedUriStringArray) {
                decodedUriStringArray[key] = decodeURI(decodedUriStringArray[key]);
            }
            for (var i = 0; i < decodedUriStringArray.length; i++) {
                UriParamsToJSON(decodedUriStringArray[i], decodeTwice, object, call + 1);
            }
        }
        else {
            //decode data
            var pair = decodeUriParam(uriString);
            //build object recursively
            mapObjectKey(pair.key, pair.value, object);
        }
        return object;
    }
    var parameters = {};
    return UriParamsToJSON(uriString, decodeTwice, parameters, 0);
}

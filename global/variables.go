package global

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

const Version string = "0.1 Demo     "
const REQUEST_MEMBERSHIP = "5sNrp1u0fG"
const APPROVAL_MEMBERSHIP = "U7nB09u7wW"
const UPDATE_NETWORK = "FcvgFRqYrg"
const UPDATE_MEMBER = "Stl1N0rBld"

const ACTION_WELCOME = "welcome"
const ACTION_DATABASE = "database"

var ConfigFile ConfigType

const BASEConfig string = `{
    "server": {
        "internal": {
            "port":"8085",            
            "tps": 100,
            "backlog": 100,
            "segurity": {
                "ipAllowed": ["127.0.0.1","127.0.0.2","198.0.1.34"],
                "cors": "*"
            }
        },
        "external": {
            "port": "8086"
        } 
    },
    "identity": {
        "pki": {
            "private": "./key.pem",
            "public": "./cert.pem",
            "ca": ".ca.pem"
        },
        "keys": {
            "private": "@privatekey",
            "public": "@publickey"
        }
    },
    "database": {        
        "path": "./store"
    },
    "log": {
        "nivel": "DEBUG",
        "ruta": "stdout",
        "color": true,
        "megabytes":  10,
        "MaxBackups": 10,
        "diasMaximo": 60
    },
    "consumidor": [],
    "proveedor": []
}`

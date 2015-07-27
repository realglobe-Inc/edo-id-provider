// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

function calculate(issuer, username, password) {
    var plus = new RegExp("\\+", "g");
    var slash = new RegExp("/", "g");
    var shaObj = new jsSHA(issuer + "\0" + username + "\0" + password, "TEXT");
    var hash = shaObj.getHash("SHA-256", "B64");
    hash = hash.substring(0, 43); // 256 bit を 6 bit 区切りにすると 43 文字。
    hash = hash.replace(plus, "-");
    hash = hash.replace(slash, "-");
    return hash;
}

function login() {
    var ticket = location.hash.substring(1)
    var queries = query_parse(window.location.search.substring(1));
    var username
    if (queries["usernames"]) {
        if (window.JSON) {
            a = JSON.parse(queries["usernames"]);
            if (a.length > 0) {
                username = a[0];
            }
        }
    }
    var form = document.form;

    if (! form) {
        return;
    }

    if (ticket) {
        var input = document.createElement("input");
        input.type = "hidden";
        input.name = "ticket";
        input.value = ticket;
        form.appendChild(input);
    }

    if (username) {
        form.elements["username"].value = username;
    }

    document.form.onsubmit = function (e) {
        var input = document.createElement("input");
        input.type = "hidden";
        input.name = "pass_type";
        input.value = "password";
        form.appendChild(input);

        // var issuer = queries["issuer"];
        // var username = form.elements["username"].value;
        // var password = form.elements["password"].value;
        // form.elements["password"].value = calculate(issuer, username, password)
    }
}

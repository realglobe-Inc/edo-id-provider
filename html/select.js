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

function select() {
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
}

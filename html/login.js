function login() {
    var uri = "/auth/login";

    var ticket = location.hash.substring(1);
    var queries = {};
    var q = window.location.search.substring(1).split("&");
    for (var i = 0; i < q.length; i++) {
        var elem = q[i].split("=");

        var key = elem[0];
        var val = elem[1];
        if (val) {
            val = decodeURIComponent(val.replace(/\+/g, " "));
        }

        queries[key] = val;
    }

    document.write('ticket: ' + ticket + '<br/>');
    for (key in queries) {
        document.write(key + ': ' + queries[key] + '<br/>');
    }

    var username = "";
    if (queries["usernames"]) {
        var buff = JSON.parse(queries["usernames"])
        if (buff[0]) {
            username = buff[0];
        }
    }

    document.write('<form method="post" action="' + uri + '">');
    document.write('アカウント: <input type="text" name="username" size="20" value="' + username + '" /> ');
    document.write('パスワード: <input type="password" name="password" size="20" /> ');
    document.write('<input type="hidden" name="ticket" value="' + ticket + '" />');
    document.write('<input type="submit" value="ログイン" />');
    document.write('</form>');
}

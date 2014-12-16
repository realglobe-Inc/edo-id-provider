function login() {
    var uri = "/login"

    // TODO 補助情報の読み取りと反映。
    var query = window.location.search;
    if (query.length > 0) {
        uri += query.replace("&", "&amp;")
    }

    document.write('<form method="post" action="' + uri + '">');
    document.write('アカウント: <input type="text" name="username" size="20" /> ');
    document.write('パスワード: <input type="password" name="passwd" size="20" /> ');
    document.write('<input type="submit" value="ログイン" />');
    document.write('</form>');
}

<?php
class Page
{
    public string $body {
        get => call_user_func(function () { return include $this->path; });
    }
    public string $head {
        get {
            return call_user_func(function () { return include 'head.php'; });
        }
    }
    private string $path = 'body.php';
}

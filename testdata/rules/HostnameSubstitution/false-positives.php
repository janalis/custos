<?php
$label   = 'Served by ' . $_SERVER['SERVER_NAME'];
$current = $_SERVER['HTTP_HOST'];
$port    = 'a@' . $_SERVER['SERVER_PORT'];
$other   = 'a@' . $_GET['HTTP_HOST'];
$items['host'] = $_SERVER['HTTP_HOST'];
echo $_SERVER['HTTP_HOST'];

class Mailer
{
    public function safe(array $known)
    {
        if (in_array($_SERVER['HTTP_HOST'], $known, true)) {
            $this->replyHost = $_SERVER['HTTP_HOST'];
        }
    }

    public function name()
    {
        $this->title = $_SERVER['HTTP_HOST'];
        return $_SERVER['SERVER_NAME'];
    }

    public function unused()
    {
        $h = $_SERVER['HTTP_HOST'];
        return strlen($h);
    }
}

function overwritten()
{
    $base = $_SERVER['HTTP_HOST'];
    $base = 'example.org';
    return 'help@' . $base;
}

function overwrittenOuter($flag)
{
    if ($flag) {
        $base = $_SERVER['HTTP_HOST'];
    }
    $base = getenv('MAIL_DOMAIN');
    return 'help@' . $base;
}

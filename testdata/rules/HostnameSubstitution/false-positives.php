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

// The assigned value is a boolean or a number, not the host name.
$isLocalhost = $_SERVER['HTTP_HOST'] === 'localhost';
$hostLength  = strlen($_SERVER['SERVER_NAME']);
$hostMissing = empty(trim($_SERVER['HTTP_HOST']));
$hostBlank   = !$_SERVER['HTTP_HOST'];
$hostNumber  = (int) $_SERVER['HTTP_HOST'];
$hostKnown   = $_SERVER['HTTP_HOST'] && $current;

function scalars($o)
{
    $o->hostname = $_SERVER['HTTP_HOST'] !== 'a.example';
    $o->hostSize = mb_strlen(trim($_SERVER['SERVER_NAME']));
    $size = strlen($_SERVER['HTTP_HOST']);
    return 'part@' . $size;
}

// The variable holding the host is checked against a list.
function validatedVariable(array $allowed)
{
    $host = $_SERVER['HTTP_HOST'] ?? '';
    if (!in_array(($host), $allowed, true)) {
        throw new RuntimeException('unexpected host');
    }
    return 'noreply@' . $host;
}

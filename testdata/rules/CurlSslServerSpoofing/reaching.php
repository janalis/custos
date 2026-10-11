<?php
class Overwritten
{
    private $verifyPeer = false;
    private $verifyHost = 0;
    private static $shared = false;

    public function __construct(bool $verifyPeer, array $opts)
    {
        $this->verifyPeer = $verifyPeer;
        $this->verifyHost = $opts['host'] ?? 0;
        self::$shared = $opts['shared'] ? false : $opts['peer'];
    }

    public function open($ch): void
    {
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->verifyPeer);
        <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $this->verifyHost)</error>;
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, self::$shared)</error>;
    }

    public function reopen($ch, array $config): void
    {
        self::$shared = resolveVerify($config);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, self::$shared);
    }
}

class Kept
{
    private $peer = false;
    private $host = 0;
    private $loose = false;

    public function __construct($dev, $cfg, $step)
    {
        $this->peer = $this->peer ?? false;
        $this->host = $dev ? 0 : $cfg;
        $step++;
        $this->loose = $step;
    }

    public function open($ch, Overwritten $other): void
    {
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer)</error>;
        <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $this->host)</error>;
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->loose);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->$ch);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $other->verifyPeer);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $other->missing);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, Overwritten::$$ch);
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, Overwritten::$shared)</error>;
        <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $other->verifyHost)</error>;
    }
}

function overwritten($ch, array $config, $name): void
{
    $verify = false;
    $verify = resolveVerify($config);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $verify);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $$name);
    $host = 0;
    $host = 2;
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $host);
}

function alternatives($ch, array $config, $peer = false, $given = false): void
{
    $verify = resolveVerify($config);
    if ($config['dev']) {
        $verify = false;
    }
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $verify)</error>;
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $peer)</error>;
    $given = $config['peer'];
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $given);
    $host = 2;
    $host = 0;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $host)</error>;
}

class Ordered
{
    private $peer;
    private $maybe = false;
    private $late = false;

    public function __construct($ch, array $config)
    {
        if (isset($config['maybe'])) {
            $this->maybe = $config['maybe'];
        }
        $this->late = $config['late'];
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->late);
        $this->late = false;
    }

    public function restored($ch, $previous): void
    {
        $this->peer = false;
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer)</error>;
        $this->peer = $previous;
    }

    public function forced($ch, $config): void
    {
        $this->peer = $config;
        $this->peer = false;
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer)</error>;
    }

    public function branched($ch, $config): void
    {
        if ($config) {
            $this->peer = false;
        } else {
            $this->peer = $config;
        }
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer)</error>;
    }

    public function inline($ch, $config): void
    {
        $this->peer = false;
        if ($this->peer = $config) {
            return;
        }
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer)</error>;
    }

    public function nested($ch, $config): void
    {
        $this->peer = false;
        foreach ($config as $entry) {
            $this->peer = $entry;
            curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->peer);
        }
    }

    public function conditionalConstructor($ch): void
    {
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->maybe)</error>;
    }
}

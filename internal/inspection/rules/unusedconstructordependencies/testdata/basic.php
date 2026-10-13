<?php
trait Describes
{
    public function describe() { return $this->label; }
}

class Mailer
{
    use Describes;

    private $transport;
    private $logger;
    private $label;
    private $retries;
    private $hook;
    /** @Inject */
    private $clock;
    /** @var int */
    private $limit;
    protected $debug;

    public function __construct($transport, $logger, $label, $clock, $limit)
    {
        $this->transport = $transport;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->logger</weak_warning> = $logger;
        $this->logger->info('ready');
        $this->label = $label;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->retries</weak_warning> = 3;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->retries</weak_warning> = $this->retries + 1;
        $register = function () { $this->hook = true; };
        $this->clock = $clock;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->limit</weak_warning> = $limit;
        $this->debug = false;
    }

    public function send($msg)
    {
        return $this->transport->push($msg);
    }

    public static function inspect(Mailer $other)
    {
        return function () use ($other) { return $other->hook; };
    }
}

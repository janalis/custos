<?php
class Mailer
{
    private $sent = false;

    public function __construct()
    {
        $this->init();
        $this->sent = false;
    }

    private function init(): void
    {
        $this->sent = true;
    }
}

class Pipe
{
    private $queue;
    private $flag = 0;
    private $other = 1;

    public function __construct(array $q)
    {
        if ($q) {
            $this->queue = $q;
        }
        $this->queue = null;
        if (!$q) {
            $this->flag = 2;
        }
        $this->flag = 0;
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->other = 1;</weak_warning>
    }
}

<?php
class Registry
{
    private array $items = [];
    public function __construct()
    {
        $this
            ->register('txt', new \stdClass())
            ->register('xml', new \stdClass());
    }
    public function register(string $name, object $item): static
    {
        $this->items[$name] = $item;
        return $this;
    }
}
class CustomRegistry extends Registry
{
    public function __construct(?object $linker = null)
    {
        $this->register('txt', $linker ?? new \stdClass());
        $this->REGISTER('md', new \stdClass());
    }
}
class Server
{
    public $host;
    public function __construct($host)
    {
        $this->host = $host;
        $this->listen();
        $fn = function () { $this->stop(); };
    }
    public function listen() {}
    public function stop() {}
}
class TlsServer extends Server
{
    public function __construct(public $host, $port)
    {
        $this->listen();
    }
}
class PartialServer extends Server
{
    public function <error descr="__construct does not call Server::__construct().">__construct</error>($host)
    {
        $this->listen();
    }
}
class QuietServer extends Server
{
    public function <error descr="__construct does not call Server::__construct().">__construct</error>($host)
    {
        $this->host = $host;
        $f = fn () => $this->listen();
    }
}
abstract class Hook
{
    abstract public function __construct();
}

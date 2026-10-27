<?php

namespace App;

class Event
{
    private $name;

    public function __construct($name)
    {
        $this->name = $name;
    }
}

// A decorator wraps an instance of its parent instead of initialising it.
final class LazyEvent extends Event
{
    public function __construct(private parent $event) {}
}

final class TracedEvent extends Event
{
    public function __construct(private ?Event $event, $tag) {}
}

final class Renamed extends Event
{
    public function <error descr="__construct does not call Event::__construct().">__construct</error>(array $names, \Closure|Event $inner, $name) {}
}

class Command
{
    public function __construct(?string $name = null) { $this->configure(); }
    protected function configure(): void {}
    public function run(): int { return 0; }
    public function help(): string { return ''; }
    public function name(): string { return ''; }
}

final class LazyCommand extends Command
{
    private \Closure|Command $command;
    public function __construct(\Closure $factory) { $this->command = $factory; }
    public function run(): int { return $this->getCommand()->run(); }
    public function help(): string { return $this->getCommand()->help(); }
    public function name(): string { return ($this->getCommand())->name(); }
    public function getCommand(): parent { return $this->command; }
}

final class Node extends Command
{
    private ?Command $owner = null;
    public function <error descr="__construct does not call Command::__construct().">__construct</error>() {}
    public function run(): int { return $this->run2(); }
    public function run2(): int { return $this->owner?->run() ?? 0; }
}

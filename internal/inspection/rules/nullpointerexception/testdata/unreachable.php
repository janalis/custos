<?php
class Box
{
    public function open(): void {}
}

class Repo
{
    public function find(int $id): ?Box { return null; }
}

class Dispatcher
{
    public function run(Repo $r, string $k): void
    {
        switch ($k) {
            case 'x':
                break;
                $b = $r->find(1);
                $b->open();
                $r->find(2)->open();
                break;
            case 'y':
                $c = $r->find(3);
                <warning descr="Possible null dereference.">$c</warning>->open();
                $r->find(4)<warning descr="Possible null dereference.">-></warning>open();
                break;
        }
    }

    public function late(Repo $r): void
    {
        return;
        $r->find(5)->open();
    }
}

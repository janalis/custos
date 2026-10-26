<?php
class Box
{
    public function open(): void {}
}

class Factory
{
    /** @return Box|mixed|null */
    public function make() { return new Box(); }

    /** @return Box|null */
    public function maybe() { return null; }

    public function use(): void
    {
        $this->make()->open();
        $this->maybe()<warning descr="Possible null dereference.">-></warning>open();
    }
}

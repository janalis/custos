<?php
class Gate
{
    public function isOpen(): bool { return true; }
    public function count(): int { return 0; }
}

function check(Gate $g)
{
    if ($g->isOpen()) {
        pass();
    } else {
        close();
    }

    if ($g->isOpen()) { b(); } else { a(); }

    if (false !== $g->count()) { d(); } else { c(); }
}

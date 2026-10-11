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

/** @return bool */
function documentedOnly($s) { return preg_match('/a/', $s); }

class Documented
{
    /** @var bool */
    public $flag = 0;
    public bool $typed = true;

    /** @param bool $strict */
    public function run($strict, bool $enforced, $raw)
    {
        // A docblock bool is not enforced: keep the strict comparison.
        if (false !== documentedOnly($raw)) { b(); } else { a(); }
        if (false !== $strict) { b(); } else { a(); }
        if (false !== $this->flag) { b(); } else { a(); }
        // Enforced by PHP: the plain condition is equivalent.
        if ($this->typed) { b(); } else { a(); }
        if ($enforced) { b(); } else { a(); }
        if ($raw > 1) { b(); } else { a(); }
        if ((bool) $raw) { b(); } else { a(); }
        if (is_string($raw)) { b(); } else { a(); }
    }
}

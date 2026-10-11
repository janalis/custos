<?php
class Gate
{
    public function isOpen(): bool { return true; }
    public function count(): int { return 0; }
}

function check(Gate $g)
{
    if (false === $g->isOpen()) {
        close();
    } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> {
        pass();
    }

    if (($g->isOpen()) === \false) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }

    if (false === $g->count()) { c(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { d(); }
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
        if (false === documentedOnly($raw)) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === $strict) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === $this->flag) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        // Enforced by PHP: the plain condition is equivalent.
        if (false === $this->typed) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === $enforced) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === ($raw > 1)) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === (bool) $raw) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
        if (false === is_string($raw)) { a(); } <weak_warning descr="Negated condition with an else branch; swap the branches and drop the negation.">else</weak_warning> { b(); }
    }
}

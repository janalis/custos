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

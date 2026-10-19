<?php
class Link
{
    public $value;
    public function next(): ?Link { return null; }

    public function loops(?Link $a, ?Link $b, ?Link $c, ?Link $d, ?Link $e, ?Link $f, ?Link $g, ?Link $h, ?Link $i)
    {
        $t = 0;
        for (; $a; $a = $a->next()) {
            $t += $a->value;
        }
        for ($k = 0; $k < 3, $b !== null; $k++, $b = $b->next()) {
            $t += $b->value;
        }
        // The body may leave $c null before the loop expression runs.
        for (; $c; $c = <warning descr="Possible null dereference.">$c</warning>->next()) {
            if ($t > 3) {
                $c = $this->next();
                continue;
            }
        }
        // Reassigned in the body before the read.
        for (; $d; ) {
            $d = $this->next();
            $t += <warning descr="Possible null dereference.">$d</warning>->value;
        }
        // An earlier loop expression assigns it.
        for (; $e; $e = $this->next(), $t += <warning descr="Possible null dereference.">$e</warning>->value) {
        }
        // No condition, or one that proves nothing.
        for (;;) {
            $t += <warning descr="Possible null dereference.">$f</warning>->value;
            break;
        }
        for (; $t < 9; $t++) {
            $t += <warning descr="Possible null dereference.">$g</warning>->value;
        }
        // The initialiser runs before the condition.
        for ($t = <warning descr="Possible null dereference.">$h</warning>->value; $h; ) {
            break;
        }
        // Reassigned then re-checked through continue.
        for (; $i; ) {
            $t += $i->value;
            if ($t > 5) {
                $i = $this->next();
                continue;
            }
            $t += $i->value;
        }
        return $t;
    }

    public function matches(?Link $a, ?Link $b, ?Link $c, ?Link $d, ?Link $e, Link $known, $loose)
    {
        $r = match ($a) {
            null => 0,
            default => $a->value,
        };
        $r += match ($b) {
            $known, $this => $b->value,
            default => 0,
        };
        $r += match ($c) {
            $known => 1,
            default => <warning descr="Possible null dereference.">$c</warning>->value,
        };
        $r += match ($d) {
            $known, null => <warning descr="Possible null dereference.">$d</warning>->value,
            $loose => $d->value,
            default => 2,
        };
        $r += match ($e) {
            null => 0,
            default => [$e = $this->next(), <warning descr="Possible null dereference.">$e</warning>->value][1],
        };
        return $r;
    }
}

function firstOf(?Link ...$links): ?Link
{
    return $links[0];
}

<?php
enum Mode { case Idle; }

class Tint {}

class Box
{
    public static function draw(?bool $checked, Mode $mode = Mode::Idle, ?Tint $tint = null): string { return ''; }
    public static function paint(?bool $checked, Tint $tint = new Tint()): string { return ''; }
}

class Cell
{
    public $value;
    public ?Cell $previous = null;
    public function next(): ?Cell { return null; }
    public function arrangement(): ?int { return null; }
    public function background(): ?string { return null; }
    public function hop(Cell $c): ?Cell { return $c->previous; }

    // Named arguments reach the parameter of that name, not the one at
    // their position.
    public function named(?Tint $tint, ?bool $on)
    {
        $a = Box::draw($on, tint: $tint);
        $b = Box::paint($on, tint: <warning descr="Possible null dereference.">$tint</warning>);
        $c = Box::draw($on, unknown: $tint);
        return $a . $b . $c;
    }

    // An earlier `if (c) { exit }` guards the statements after it. The
    // nullsafe checks below do not stop the walk, unlike a plain null check.
    public function earlyExit(?Cell $a, ?Cell $b, ?Cell $c, ?Cell $d, ?Cell $e, ?Cell $f, ?Cell $g)
    {
        if (null === $a?->arrangement()) {
            return;
        }
        $a->value = 1;

        if (null === $b?->arrangement()) {
            echo 'no exit';
        }
        <warning descr="Possible null dereference.">$b</warning>->value = 1;

        if (null === $c?->arrangement()) {
            $c = $this->next();
            throw new \LogicException();
        }
        $c->value = 1;

        if (null === $d?->arrangement()) {
            return;
        }
        $d = $this->next();
        <warning descr="Possible null dereference.">$d</warning>->value = 1;

        foreach ([1, 2] as $i) {
            if (null === $e?->arrangement()) {
                continue;
            }
            $e->value = $i;
        }

        if (null === $g?->arrangement()) {
            $g = $this->next();
            return;
        }
        $g->value = 1;

        if (null === $e?->arrangement() || null === $f?->arrangement()) {
            exit(1);
        }
        $e->value = $f->value;
    }

    // A continue that does not go back through the guarding condition.
    public function outerGuard(?Cell $w, int $k)
    {
        if (null !== $w?->arrangement()) {
            foreach ([1, 2] as $i) {
                if ($i > $k) {
                    $w = $this->next();
                    continue;
                }
                <warning descr="Possible null dereference.">$w</warning>->value = 1;
            }
        }
    }

    // An assignment inside a closure may run before the use.
    public function byReference(?Cell $n)
    {
        if (null !== $n?->arrangement()) {
            $f = function () use (&$n) {
                $n = $this->next();
            };
            $f();
            <warning descr="Possible null dereference.">$n</warning>->value = 1;
        }
    }

    // match (true) arms run only when their condition is true.
    public function arms(?Cell $row, ?Cell $other, ?Cell $third, int $k)
    {
        $x = match (true) {
            null !== $row?->background() => $row->background(),
            default => null,
        };
        $y = match (true) {
            null !== $other?->background(), $k > 1 => <warning descr="Possible null dereference.">$other</warning>->value,
            default => null,
        };
        $z = match ($k) {
            1 => <warning descr="Possible null dereference.">$third</warning>->value,
            default => null,
        };
        return [$x, $y, $z];
    }

    // A loop condition still holds after a branch that reassigns the
    // variable and then jumps back to that condition. Each local is first
    // assigned inside its loop, so the walk starts after the loop check.
    public function loops(Cell $start, int $k)
    {
        $closer = $start->previous;
        while (null !== $closer) {
            if ($closer->value > 1) {
                $closer = $this->hop($closer);
                continue;
            }
            $closer->value = 1;
            $closer = $closer->previous;
        }

        $p = $start->previous;
        while (null !== $p) {
            if ($k > 1) {
                $p = $this->next();
                break;
            }
            $p->value = 1;
        }

        $q = $start->previous;
        while (null !== $q) {
            switch ($k) {
                case 1:
                    $q = $this->next();
                    break;
            }
            <warning descr="Possible null dereference.">$q</warning>->value = 1;
        }

        $r = $start->previous;
        while (null !== $r) {
            foreach ([1] as $i) {
                if ($i > $k) {
                    $r = $this->next();
                    continue 2;
                }
            }
            <warning descr="Possible null dereference.">$r</warning>->value = 1;
        }

        $u = $start->previous;
        while (null !== $u) {
            $u = $this->next();
            <warning descr="Possible null dereference.">$u</warning>->value = 1;
            return;
        }

        $t = $start->previous;
        foreach ([1] as $i) {
            while (null !== $t) {
                if ($i > $k) {
                    $t = $this->next();
                    continue 2;
                }
                $t->value = 1;
            }
        }

        $m = $start->previous;
        while (null !== $m) {
            if ($k > 1) {
                $m = $this->next();
                if ($k > 2) {
                    return;
                } else {
                    return;
                }
            }
            <warning descr="Possible null dereference.">$m</warning>->value = 1;
        }

        $s = $start->previous;
        while (null !== $s) {
            if ($k > 1) {
                $s = $this->next();
            }
            <warning descr="Possible null dereference.">$s</warning>->value = 1;
        }
    }
}

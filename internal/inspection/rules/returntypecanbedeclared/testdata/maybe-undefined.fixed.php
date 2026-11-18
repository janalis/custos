<?php
class Finder
{
    // Returned while possibly unassigned: the call yields null.
    public function regexp($t): ?string
    {
        switch ($t) {
            case 'x':
                $r = 'X';
                break;
            case 'y':
                $r = 'Y';
                break;
        }
        return $r;
    }

    public function maybe($t): ?string
    {
        if ($t) {
            $r = 'X';
        }
        return $r;
    }

    // Certainly assigned, parameters, other binders: inference decides
    // (extract(), `$$n =`, parse_str() and include may set $r: unknown).
    public function always($t): string
    {
        $r = 'Y';
        if ($t) {
            $r = 'X';
        }
        return $r;
    }

    public function param(string $r): string
    {
        if ($r === '') {
            $r = 'X';
        }
        return $r;
    }

    public function compound($t): string
    {
        if ($t) {
            $r = 'X';
        }
        $r .= 'Y';
        return $r;
    }

    public function element($t): array
    {
        if ($t) {
            $r = [];
        }
        $r[] = 1;
        return $r;
    }

    public function byRef($t, $s): array
    {
        if ($t) {
            $m = [];
        }
        preg_match('/x/', $s, $m);
        return $m;
    }

    public function extracted($t, array $vars)
    {
        if ($t) {
            $r = 'X';
        }
        extract($vars);
        return $r;
    }

    public function variable($t, $n)
    {
        if ($t) {
            $r = 'X';
        }
        $$n = 'Z';
        return $r;
    }

    public function included($t)
    {
        if ($t) {
            $r = 'X';
        }
        include 'defaults.php';
        return $r;
    }

    public function parsed($t, $q)
    {
        if ($t) {
            $r = 'X';
        }
        parse_str($q);
        return $r;
    }

    public function destructured($t, array $pair): string
    {
        if ($t) {
            $r = 'X';
        }
        [$r] = ['Y'];
        return $r;
    }

    // A reference alias may be written through its other name: unknown.
    public function byRefAssign($t, string $s)
    {
        if ($t) {
            $r = 'X';
        }
        $r = &$s;
        return $r;
    }
}

<?php
class Finder
{
    // Returned while possibly unassigned: the call yields null.
    public function <weak_warning descr="Declare ': ?string' as the return type.">regexp</weak_warning>($t)
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

    public function <weak_warning descr="Declare ': ?string' as the return type.">maybe</weak_warning>($t)
    {
        if ($t) {
            $r = 'X';
        }
        return $r;
    }

    // Certainly assigned, parameters, other binders: inference decides
    // (extract(), `$$n =` and parse_str() may set $r: unknown).
    public function <weak_warning descr="Declare ': string' as the return type.">always</weak_warning>($t)
    {
        $r = 'Y';
        if ($t) {
            $r = 'X';
        }
        return $r;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">param</weak_warning>(string $r)
    {
        if ($r === '') {
            $r = 'X';
        }
        return $r;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">compound</weak_warning>($t)
    {
        if ($t) {
            $r = 'X';
        }
        $r .= 'Y';
        return $r;
    }

    public function <weak_warning descr="Declare ': array' as the return type.">element</weak_warning>($t)
    {
        if ($t) {
            $r = [];
        }
        $r[] = 1;
        return $r;
    }

    public function <weak_warning descr="Declare ': array' as the return type.">byRef</weak_warning>($t, $s)
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

    public function <weak_warning descr="Declare ': string' as the return type.">included</weak_warning>($t)
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

    public function <weak_warning descr="Declare ': string' as the return type.">destructured</weak_warning>($t, array $pair)
    {
        if ($t) {
            $r = 'X';
        }
        [$r] = ['Y'];
        return $r;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">byRefAssign</weak_warning>($t, string $s)
    {
        if ($t) {
            $r = 'X';
        }
        $r = &$s;
        return $r;
    }
}

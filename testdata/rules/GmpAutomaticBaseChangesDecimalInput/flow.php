<?php

class Limits
{
    const DIGITS = '017';

    public function fromLocal(array $rows)
    {
        $raw = '024';
        $other = 1;
        $other++;
        $keep = [&$other];
        parse_str('a=1', $out);
        $print = function () use ($raw, &$other) {
            return $raw;
        };

        return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init($raw)</warning>;
    }

    public function lateBound()
    {
        return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init(static::DIGITS)</warning>;
    }

    public function explicitZero()
    {
        $raw = '024';

        return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init($raw, 0)</warning>;
    }

    // A nested class body is another scope: its `++` leaves the local alone.
    public function nested()
    {
        $raw = '010';
        $helper = new class {
            public function bump()
            {
                $raw = 1;
                $raw++;
            }
        };

        return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init($raw)</warning>;
    }
}

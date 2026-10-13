<?php
class Desk
{
    const NAME = 'Desk';
    public function open() {}
    protected function guarded() {}
    private function secret() {}
}

function shapes(Desk $d, string $m)
{
    return [
        is_callable(42),
        is_callable(<<<'EOT'
Desk::open
EOT),
        is_callable([$d, $m]),
        is_callable([$d, '']),
        is_callable([$d, <<<'EOT'
open
EOT]),
        is_callable([Desk::NAME, 'open']),
        is_callable(new Desk()),
    ];
}

$desk = new Desk();
is_callable(<warning descr="Method 'guarded' is not public, so the callback cannot be invoked from outside.">[$desk, 'guarded']</warning>);

class Holder
{
    public function probe()
    {
        $plain = new class {
            public function t(Desk $d) { return is_callable(<warning descr="Method 'guarded' is not public, so the callback cannot be invoked from outside.">[$d, 'guarded']</warning>); }
        };
        $child = new class extends Desk {
            public function t() { return is_callable([new Desk(), 'guarded']); }
            public function u() { return is_callable(<warning descr="Method 'secret' is not public, so the callback cannot be invoked from outside.">[new Desk(), 'secret']</warning>); }
        };
        $other = new class extends Holder {
            public function t() { return is_callable(<warning descr="Method 'guarded' is not public, so the callback cannot be invoked from outside.">[new Desk(), 'guarded']</warning>); }
        };
        return [$plain, $child, $other];
    }
}

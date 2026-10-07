<?php
namespace Plugins {
    class Hook
    {
        public function fire() {}
        private function secret() {}
    }

    // namespace-local is_callable with different semantics
    function is_callable($cb) { return true; }

    function probe()
    {
        return [
            is_callable(['Plugins\Hook', 'secret']),
            Registry\is_callable('Plugins\Hook::fire'),
            \Is_Callable(<warning descr="Method 'fire' is not static but is referenced without an object.">'Plugins\Hook::fire'</warning>),
        ];
    }
}

namespace {
    use function Plugins\is_callable;

    function imported()
    {
        return is_callable('Plugins\Hook::fire');
    }
}

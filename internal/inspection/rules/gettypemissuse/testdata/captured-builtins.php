<?php
namespace Checks {
    use function Polyfill\is_int;

    function kind($x) {
        return <warning descr="Use '\is_int($x)' instead.">gettype($x) === 'integer'</warning>;
    }
}

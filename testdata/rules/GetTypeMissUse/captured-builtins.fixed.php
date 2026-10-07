<?php
namespace Checks {
    use function Polyfill\is_int;

    function kind($x) {
        return \is_int($x);
    }
}

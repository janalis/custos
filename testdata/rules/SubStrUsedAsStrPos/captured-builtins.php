<?php
namespace Routing {
    function strpos($haystack, $needle) { return 0; }

    function match($uri, $base) {
        return <weak_warning descr="Use '\strpos($uri, $base) === 0' instead.">substr($uri, 0, strlen($base)) === $base</weak_warning>;
    }
}

namespace Imported {
    use function Routing\strpos as mb_strpos;

    function match($uri, $base) {
        return <weak_warning descr="Use '\mb_strpos($uri, $base) === 0' instead.">mb_substr($uri, 0, mb_strlen($base)) === $base</weak_warning>;
    }
}

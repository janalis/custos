<?php
namespace Routing {
    function strpos($haystack, $needle) { return 0; }

    function match($uri, $base) {
        return \strpos($uri, $base) === 0;
    }
}

namespace Imported {
    use function Routing\strpos as mb_strpos;

    function match($uri, $base) {
        return \mb_strpos($uri, $base) === 0;
    }
}

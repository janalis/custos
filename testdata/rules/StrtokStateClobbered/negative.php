<?php
namespace Negative0 { strtok($a,',');strtok(','); }
namespace Negative1 { strtok($a,',');strtok($b,':'); }
namespace Negative2 { strtok($a,',');unknown();strtok($b,':');strtok(','); }
namespace Negative3 { strtok($a,',');strtok(',');strtok($b,':');strtok(','); }

namespace LeadingToken {
    // The first call only takes a leading token; the continuation belongs to the newer sequence.
    function parse(string $url, string $file) {
        $scheme = strtok($url, ':');
        $name = strtok($file, '.');
        $ext = strtok('.');
        return [$scheme, $name, $ext];
    }
    function loop($x, $y) {
        $a = strtok($x, ',');
        $b = strtok($y, ';');
        while (($c = strtok(';')) !== false) { echo $c; }
    }
}

namespace Barriers {
    function sameString($x) { $a = strtok($x, ','); $b = strtok($x, ';'); $c = strtok(','); }
    function callAfterSecond($x, $y) { $a = strtok($x, ','); $b = strtok($y, ';'); helper(); $c = strtok(','); }
    function branchStart($x, $y, $q) { strtok($x, ','); if ($q) { strtok($y, ':'); } strtok(','); }
    function elseBranch($x, $y, $q) { strtok($x, ','); strtok($y, ':'); if ($q) { echo 1; } else { strtok(','); } }
    function callingHeader($x, $y) { strtok($x, ','); strtok($y, ':'); while (more()) { strtok(','); } }
    function callingInit($x, $y) { strtok($x, ','); strtok($y, ':'); for ($i = init(); $i < 3; $i++) { strtok(','); } }
    function otherConstruct($x, $y) { strtok($x, ','); strtok($y, ':'); do { strtok(','); } while (false); }
    function suspended($x, $y) { strtok($x, ','); strtok($y, ':'); yield 1; strtok(','); }
    function methodBetween($x, $y, $o) { strtok($x, ','); strtok($y, ':'); $o->run(); strtok(','); }
    function closureBetween($x, $y) { strtok($x, ','); strtok($y, ':'); $f = function () {}; strtok(','); }
    function onlyOneStart($y, $q) { if ($q) { strtok($y, ':'); strtok(','); } }
}

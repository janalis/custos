<?php
function falsyExit($d) { $fs = scandir($d); if (!$fs) { return -1; } $n = 0; foreach ($fs as $f) { $n++; } return $n; }
function arrayBranch($d) { $fs = scandir($d); $n = 0; if (is_array($fs)) { foreach ($fs as $f) { $n++; } } return $n; }
function truthyBranch($d) { $fs = scandir($d); $n = 0; if ($fs) { foreach ($fs as $f) { $n++; } } return $n; }
function notArrayExit($d) { $fs = scandir($d); if (!\is_array($fs)) { return -1; } $n = 0; foreach ($fs as $f) { $n++; } return $n; }
function looseExit($d) { $fs = scandir($d); if ($fs == false) { return -1; } $n = 0; foreach ($fs as $f) { $n++; } return $n; }
function yodaLooseBranch($d) { $fs = scandir($d); $n = 0; if (false != $fs) { foreach ($fs as $f) { $n++; } } return $n; }
function emptyExit($d) { $fs = scandir($d); if (empty($fs)) { return -1; } $n = 0; foreach ($fs as $f) { $n++; } return $n; }
function eitherExit($d) { $fs = scandir($d); if (!is_array($fs) || count($fs) === 0) { return; } echo 1; foreach ($fs as $f) { echo $f; } }
function strictEitherExit($d) { $fs = scandir($d); if ($fs === false or $fs === []) { return; } foreach ($fs as $f) { echo $f; } }
function bothBranch($d, $c) { $fs = scandir($d); if ($c && $fs) { foreach ($fs as $f) { echo $f; } } }
function bothOperands($d) { $fs = scandir($d); if ($fs || is_array($fs)) { foreach ($fs as $f) { echo $f; } } }
function neitherExit($d) { $fs = scandir($d); if (!$fs and empty($fs)) { return; } foreach ($fs as $f) { echo $f; } }
function elseBranch($d) { $fs = scandir($d); if (!$fs) { echo 0; } else { foreach ($fs as $f) { echo $f; } } }
function elseifBranch($d, $c) { $fs = scandir($d); if ($c) { echo 0; } elseif ($fs) { foreach ($fs as $f) { echo $f; } } }
function afterElseif($d, $c) { $fs = scandir($d); if ($c) { echo 0; } elseif (!$fs) { echo 1; } else { foreach ($fs as $f) { echo $f; } } }
function elseExit($d) { $fs = scandir($d); if ($fs) { echo 1; } else { return; } foreach ($fs as $f) { echo $f; } }
function outerExit($d, $c) { $fs = scandir($d); if (!$fs) { throw new RuntimeException($d); } if ($c) foreach ($fs as $f) { echo $f; } }
function loopExit($d, $ds) { $fs = scandir($d); if (!$fs) { return; } while ($ds) { $ds = next($ds); foreach ($fs as $f) { echo $f; } } }

function retried_listing($directory, $separator) {
    $attempts = 0;
    again:
    try {
        $names = scandir($directory);
        if (!is_array($names)) {
            return 0;
        }
        foreach ($names as $name) {
            echo $directory . $separator . $name;
        }
    } catch (RuntimeException $failure) {
        if (++$attempts < 3) {
            goto again;
        }
        throw $failure;
    }
    return 1;
}

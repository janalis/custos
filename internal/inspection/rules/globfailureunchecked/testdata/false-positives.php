<?php
namespace Local {
    function glob($pattern) { return []; }
    function count($value) { return 0; }
    foreach (glob('*.txt') as $p) { echo $p; }
    echo count(glob('*.txt'));
}

namespace App {
    function falsyReturn($dir) { $paths = glob($dir); if (!$paths) { return; } foreach ($paths as $p) { echo $p; } }
    function arrayBranch($dir) { $paths = glob($dir); if (is_array($paths)) { foreach ($paths as $p) { echo $p; } } }
    function truthyBranch($dir) { $paths = glob($dir); if ($paths) { foreach ($paths as $p) { echo $p; } } }
    function emptyReturn($dir) { $paths = glob($dir); if (empty($paths)) { return []; } return array_map('basename', $paths); }
    function notArrayReturn($dir) { $paths = glob($dir); if (!is_array($paths)) { return; } foreach ($paths as $p) { echo $p; } }
    function looseReturn($dir) { $paths = glob($dir); if ($paths == false) { return; } foreach ($paths as $p) { echo $p; } }
    function boolReturn($dir) { $paths = glob($dir); if (is_bool($paths)) { return 0; } return count($paths); }
    function predicateCompared($dir) { $paths = glob($dir); if (is_array($paths) === false) { return; } foreach ($paths as $p) { echo $p; } }
    function elseBranch($dir) { $paths = glob($dir); if (!$paths) { echo 'none'; } else { foreach ($paths as $p) { echo $p; } } }
    function elseIfBranch($dir, $quiet) { $paths = glob($dir); if ($quiet) { return; } elseif (\is_iterable($paths)) { foreach ($paths as $p) { echo $p; } } }
    function laterStatement($dir) { $paths = glob($dir); if (!$paths) { return 0; } $seen = 0; echo $seen; return count($paths); }
    function conjunction($dir) { $paths = glob($dir); if ($paths && count($paths) > 1) { foreach ($paths as $p) { echo $p; } } }
    function ternary($dir) { $paths = glob($dir); return $paths ? count($paths) : 0; }
    function ternaryElse($dir) { $paths = glob($dir); return empty($paths) ? [] : array_values($paths); }
    function shortCircuit($dir) { $paths = glob($dir); return $paths === false || count($paths) === 0; }
    function offsetSet($dir) { $paths = glob($dir); if (isset($paths[0])) { foreach ($paths as $p) { echo $p; } } }
    function leaves($dir) { $paths = glob($dir); $paths or exit(1); foreach ($paths as $p) { echo $p; } }
    function throws($dir) { $paths = glob($dir); $paths === false and throw new \RuntimeException('glob'); return count($paths); }
    function fallback($dir) { return count(glob($dir) ?: []); }
    function cast($dir) { return count((array) glob($dir)); }
}

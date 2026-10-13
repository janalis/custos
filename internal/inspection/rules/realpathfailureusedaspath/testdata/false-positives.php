<?php
function negatedExit(string $dir) { $root = realpath($dir); if (!$root) { return null; } return is_file($root . '/a.json'); }
function stringBranch(string $dir) { $root = realpath($dir); if (is_string($root)) { return is_file($root . '/a.json'); } return null; }
function truthyBranch(string $dir) { $root = realpath($dir); if ($root) { return is_file($root . '/a.json'); } return null; }
function notStringExit(string $dir) { $root = realpath($dir); if (!is_string($root)) { return null; } return is_file($root . '/a.json'); }
function ternaryArm(string $dir) { $root = realpath($dir); return $root === false ? null : is_file($root . '/a.json'); }
function disjunctExit(string $dir) { $root = realpath($dir); if (!$root || !is_dir($root)) { return null; } return is_file($root . '/a.json'); }
function conjunctBranch(string $dir) { $root = realpath($dir); if ($root && is_dir($root)) { return is_file($root . '/a.json'); } return null; }
function looseExit(string $dir) { $root = realpath($dir); if ($root == false) { return null; } return is_file($root . '/a.json'); }
function emptyExit(string $dir) { $root = realpath($dir); if (empty($root)) { return null; } return is_file($root . '/a.json'); }
function orDie(string $dir) { $root = realpath($dir) or die('missing'); return is_file($root . '/a.json'); }
function inCondition(string $dir) { if (!($root = realpath($dir))) { return null; } return is_file($root . '/a.json'); }
function lateCheck(string $dir) { $root = realpath($dir); $file = $root . '/a.json'; if ($root === false) { return null; } return is_file($file); }

function retried_resolution($directory, $leaf) {
    $attempts = 0;
    again:
    try {
        $resolved = realpath($directory);
        if (!is_string($resolved)) {
            return null;
        }
        return file_get_contents($resolved . '/' . $leaf);
    } catch (RuntimeException $failure) {
        if (++$attempts < 3) {
            goto again;
        }
        throw $failure;
    }
}

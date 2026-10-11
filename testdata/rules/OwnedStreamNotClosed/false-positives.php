<?php
final class Reader
{
    public $handle;
    public static $pool = [];
    public function __construct($handle) { $this->handle = $handle; }
    public function __destruct() { fclose($this->handle); }

    public function keep($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } $this->handle = $h; return fgets($h); }
    public function pool($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } self::$pool[] = $h; return fgets($h); }
    public function adopt($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return $this->wrap($h); }
    private function wrap($h) { return new self($h); }
}

// The return expression closes the handle itself.
function closes($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fclose($h); }
function readsThenCloses($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fgets($h) . (fclose($h) ? '' : '!'); }

// The caller receives the live handle.
function wraps($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return new Reader($h); }
function pair($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return [$h, filesize($path)]; }
function keyed($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return ['handle' => $h]; }
function orNull($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return $h ?: null; }
function checked($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return is_resource($h) ? $h : null; }
function lazy($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return function () use ($h) { return fgets($h); }; }
function lazyArrow($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fn () => fgets($h); }
function custom($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return consume($h); }
function spread($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fgets(...$h); }

// Ownership left the function before the return.
function atShutdown($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } register_shutdown_function('fclose', $h); return fgets($h); }
function global_($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } $GLOBALS['h'] = $h; return fgets($h); }
function aliased($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } $copy = $h; return fgets($h); }
function rebound($path, &$slot) { $h = fopen($path, 'rb'); if ($h === false) { return false; } $h = &$slot; return fgets($h); }
function numbered($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return (int) $h + strlen(fgets($h)); }

<?php
namespace App;

function recompute(array $t) { return 1; }
function strlen($s) { return 0; }

// S4: side effects are never reordered
if (fwrite($log, $line) !== false && $verbose) {}
if (rename($tmp, $target) || $force) {}
if (mkdir($dir) || $dryRun) {}
if (session_start() && $user) {}
if (preg_match('/^v(\d+)/', $tag, $parts) && $strict) {}
if (recompute($totals) > 0 && $dirty) {}       // user function: impure
if (strlen($name) > 0 && $ok) {}               // namespaced user strlen shadows the built-in
if (\strlen(file_get_contents($path)) && $ok) {}
if ($handler($x) && $ok) {}                    // dynamic call
if (call_user_func($cb) && $ok) {}
if ((include $file) && $ok) {}
if (unknown_fn() && $ok) {}
if ($ok && fclose($h)) {}                      // costlier second operand: never reported anyway
if (\strlen($a) > 1 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$b == 1</weak_warning> && fflush($h)) {} // only neighbours are compared

// method, static and constructor calls may have side effects
if ($repo->save($e) || $quiet) {}
if ($repo?->flush() && $ok) {}
if (Cache::clear() || $quiet) {}
if (new Lock($path) && $ok) {}

// pure built-ins stay reorderable
if (\count($rows) > 0 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$ready</weak_warning>) {}
if (\ctype_digit($v) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$v</weak_warning>) {}
if (\file_exists($p) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$skip</weak_warning>) {}
if (\uniqid() && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$x</weak_warning> && \fwrite($h, $l)) {}
if (\preg_match('/a/', fn() => fwrite($h, 'x')) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$y</weak_warning>) {}

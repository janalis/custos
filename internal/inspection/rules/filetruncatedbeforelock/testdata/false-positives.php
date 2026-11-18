<?php
function good($path) { $h = fopen($path, 'cb'); if ($h === false) { return; } if (flock($h, LOCK_EX)) { ftruncate($h, 0); } fclose($h); }
function unlockOnly($path) { $h = fopen($path, 'w'); fwrite($h, 'x'); flock($h, LOCK_UN); fclose($h); }
function sharedOnly($path) { $h = fopen($path, 'w'); flock($h, LOCK_SH | LOCK_NB); fclose($h); }
function dynamicOperation($path, $operation) { $h = fopen($path, 'w'); flock($h, $operation); fclose($h); }
function partlyDynamic($path, $extra) { $h = fopen($path, 'w'); flock($h, LOCK_EX | $extra); fclose($h); }
function lockOnly($path) { $h = fopen($path, 'w'); flock($h, LOCK_EX); $this->locks[] = $h; }

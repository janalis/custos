<?php
function conditionalRejection($s) {
curl_setopt($s, CURLOPT_RETURNTRANSFER, true);
if (!($result = curl_exec($s))) { if ($result === false) { throw new RuntimeException(); } }
}
function successfulValue($s) {
curl_setopt($s, CURLOPT_RETURNTRANSFER, true);
if (!!curl_exec($s)) { return false; }
}
function rejectEitherPath($s, $flag) {
curl_setopt($s, CURLOPT_RETURNTRANSFER, true);
if (!<warning descr="Compare cURL transport failure strictly with false.">curl_exec($s)</warning>) { if ($flag) { return false; } else { throw new RuntimeException(); } }
}

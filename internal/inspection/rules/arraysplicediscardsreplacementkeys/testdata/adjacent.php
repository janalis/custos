<?php
function local() {
    $set = function () use (&$a) { $a["name"] = "stored"; };
    $set();
    $a = ["id" => 1];
    <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($a, 0, 0, ["name" => "Ada"])</warning>;
    echo $a["name"];
}
function shared() {
    global $g;
    $g = ([]);
    <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($g, 0, 0, ["name" => "Ada"])</warning>;
    echo $g["name"];
}
function inBranch($c) {
    $a = [];
    if ($c) {
        <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($a, 0, 0, ["name" => "Ada"])</warning>;
        echo $a["name"];
    }
}
function afterOtherWork() {
    $a = [];
    $b = [];
    fill();
    <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($a, 0, 0, ["name" => "Ada"])</warning>;
    echo $a["name"];
}

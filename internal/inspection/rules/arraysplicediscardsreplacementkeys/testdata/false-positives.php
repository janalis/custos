<?php
function viaClosure() {
    $a = [];
    $set = function () use (&$a) { $a["name"] = "stored"; };
    $set();
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function fill() { global $g; $g["name"] = "global"; return "Ada"; }
function viaGlobal() {
    global $g;
    $g = [];
    fill();
    array_splice($g, 0, 0, ["name" => "Ada"]);
    echo $g["name"];
}
function viaArgument() {
    global $g;
    $g = [];
    array_splice($g, 0, 0, ["name" => fill()]);
    echo $g["name"];
}
function viaAlias() {
    $a = [];
    $b = &$a;
    $b["name"] = "alias";
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaBranch($c) {
    $a = [];
    if ($c) {
        $a = ["name" => "kept"];
    }
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaLaterUse($c) {
    $a = [];
    if ($c) {
        array_splice($a, 0, 0, ["name" => "Ada"]);
        echo $a["name"];
    }
    return $a;
}
function viaInclude() {
    $a = [];
    include __DIR__ . '/fill.php';
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaExtract(array $vars) {
    $a = [];
    extract($vars);
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaParameter(array &$a) {
    $a = [];
    fill();
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
$viaCapture = function () use (&$a) {
    $a = [];
    fill();
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
};
$top = [];
fill();
array_splice($top, 0, 0, ["name" => "Ada"]);
echo $top["name"];
function viaReferenceBinding(array &$other) {
    $a = &$other;
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaCompound() {
    $a = [];
    $a += ["name" => "kept"];
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}
function viaDeclaration() {
    $a = [];
    global $a;
    array_splice($a, 0, 0, ["name" => "Ada"]);
    echo $a["name"];
}

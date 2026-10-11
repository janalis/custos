<?php
$db=new mysqli();$s=$db->prepare("SELECT ? /* ? */ , '?'");$s->bind_param("i",$id);$s->bind_param("i",...$ids);$unknown=$db->prepare($query);$unknown->bind_param("i",$id);$vendor=$db->prepare("SELECT $$?$$");$vendor->bind_param("i",$id);$s->execute();
function direct(mysqli_stmt $s){$s->bind_param("i",$id);}
function vendorComments(mysqli $db, $a, $b) {
    $dash = $db->prepare("SELECT 5--?");
    $dash->bind_param("i", $a);
    $executable = $db->prepare("SELECT ? /*! + ? */");
    $executable->bind_param("ii", $a, $b);
    $versioned = $db->prepare("SELECT ? /*M!100100 + ? */");
    $versioned->bind_param("ii", $a, $b);
    $hint = $db->prepare("SELECT /*+ MAX_EXECUTION_TIME(?) */ ?");
    $hint->bind_param("i", $a);
}

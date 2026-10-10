<?php
$db=new mysqli();$s=$db->prepare("SELECT ? /* ? */ , '?'");$s->bind_param("i",$id);$s->bind_param("i",...$ids);$unknown=$db->prepare($query);$unknown->bind_param("i",$id);$vendor=$db->prepare("SELECT $$?$$");$vendor->bind_param("i",$id);$s->execute();
function direct(mysqli_stmt $s){$s->bind_param("i",$id);}

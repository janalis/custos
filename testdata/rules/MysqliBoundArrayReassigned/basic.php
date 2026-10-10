<?php
$db=new mysqli();$s=$db->prepare("SELECT ?");$row=["id"=>1];$s->bind_param("i",$row["id"]);<warning descr="Update the bound array element without replacing its array.">$row=["id"=>2]</warning>;$s->execute();

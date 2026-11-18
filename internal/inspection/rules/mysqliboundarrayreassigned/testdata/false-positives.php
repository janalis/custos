<?php
$db=new mysqli();$s=$db->prepare("SELECT ?");$row=["id"=>1];$s->bind_param("i",$row["id"]);$row["id"]=2;$s->execute();$s->execute();$s->bind_param("i",$id);$row=["id"=>3];$s->execute();$s->fetch();

<?php
$pdo=new PDO("sqlite::memory:");$s=$pdo->prepare("SELECT ?");<warning descr="Expand list values into separate SQL placeholders.">$s->execute([[1,2]])</warning>;

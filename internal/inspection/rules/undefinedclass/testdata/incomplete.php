<?php
// Without a complete world (Composer packages installed and indexed), a
// missing class may be declared by code the index never saw.
$user = new \App\Model\Usr();
class Admin extends \App\BaseUser {}

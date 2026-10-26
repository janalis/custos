<?php
namespace App;

class UserRepository extends \Acme\Orm\EntityRepository {}
class BaseRepository extends \Acme\Orm\EntityRepository {}
class AuditedRepository extends BaseRepository {}
class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">ArchiveRepository</weak_warning> extends AuditedRepository {}
class Rows extends \RecursiveArrayIterator {}

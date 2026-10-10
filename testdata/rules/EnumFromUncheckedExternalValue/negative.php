<?php
namespace Negative0 { enum E:string{case A='a';} E::tryFrom($_GET['x']); }
namespace Negative1 { enum E:string{case A='a';} E::from('a'); }
namespace Negative2 { enum E:string{case A='a';} try{E::from($_GET['x']);}catch(\ValueError $e){} }
namespace Negative3 { enum E{case A;} E::from($_GET['x']); }
namespace Negative4 { Unknown::from($_GET['x']); }
namespace Negative5 { class E{static function from($x){}} E::from($_GET['x']); }
namespace Negative6 { enum E:string{case A='a';} E::from($x); }
namespace Negative7 { enum E:string{case A='a';}E::from($value['x']); }
namespace Negative8 { enum E:string{case A='a';}E::from(getInput()['x']); }

namespace AuditMembership {enum E:string{case A='a';}if(\in_array($_GET['x'],['a'],true)){E::from($_GET['x']);}}

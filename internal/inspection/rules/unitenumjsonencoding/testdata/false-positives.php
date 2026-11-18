<?php
namespace Negative0 { json_encode($x); }
namespace Negative1 { enum E:string{case A='a';}json_encode(E::A); }
namespace Negative2 { enum E implements \JsonSerializable{case A;function jsonSerialize():mixed{return 'a';}}json_encode(E::A); }
namespace Negative3 { json_encode('x'); }
namespace Negative4 { enum E{case A;}json_encode(E::A->name); }
namespace Negative5 { strlen('x'); }

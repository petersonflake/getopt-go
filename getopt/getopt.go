//
//Parse command-line options
//
//This package provides simple long and short option parsing.
//Short, i.e. single byte, options are recognized by being
//proceeded by '-', long options by "--".  It provides the
//
//following option types:
//Flag:  Either true or false.  Set to true if passed.  Passing
//the short flag with '+' will set the flag to false, as will
//setting the long option equal to false.  E.g., the arguments
//"+f" and "--force=False" will both set the flag to false.
//"-f" and "--force" and "--force=True" will all set the flag
//to true.  This is to facilitate shell scripting, since
//options can be set by default and overridden
//
//OptArg:  Takes a single argument.  Can be set like
//--file=some_file.txt or --file some_file.txt using long
//options, or -fsome_file.txt or -f some_file.txt all set
//that option to some_file.txt.  Subsequent occurrences of
//the option will overwrite the previous value
//
//OptVec:  Takes one or more arguments.  Can be set like
//OptArg, except that multiple occurrences will append
//to the array of arguments for the option.  Can be used
//for, e.g., processing multiple files
//
//OptCount:  Returns the number of times it has been passed,
//or a number passed directly.  For example:
//"-vv", "--verbose --verbose" and "--verbose=2" all accomplish
//the same thing.  If the short option is negated, "+v" then
//the value is subtracted instead of incremented
//
//Options should be created with their respective constructors, since
//this stores the option in the maps and lists, which is used for
//parsing
//
//Use ParseArgv to parse a supplies argument vector, and GetOpts to parse
//os.Args
package getopt

import(
	"errors"
	"strings"
	"fmt"
	"strconv"
	"os"
)

//Print program name, description, version and help
func PrintHelp() {
	fmt.Printf("%s - %s\n", ProgramName, ProgramVersion)
	fmt.Println(ProgramDesc)
	f := "-%c/--%-32s\t%s\n"
	for _, opt := range optByLong {
		switch opt.(type) {
		case *Flag:
			fmt.Printf(f, opt.(*Flag).Short, opt.(*Flag).Long, opt.(*Flag).Help)
		case *OptArg:
			fmt.Printf(f, opt.(*OptArg).Short, opt.(*OptArg).Long, opt.(*OptArg).Help)
		case *OptVec:
			fmt.Printf(f, opt.(*OptVec).Short, opt.(*OptVec).Long, opt.(*OptVec).Help)
		case *OptCount:
			fmt.Printf(f, opt.(*OptCount).Short, opt.(*OptCount).Long, opt.(*OptCount).Help)
		default:
			panic("Unexpected type in array of Opt by long")
		}
	}
}

//Print program name and version
func PrintVersion() {
	fmt.Printf("%s - %s\n", ProgramName, ProgramVersion)
	panic("Not implemented")
}


// A flag is either true or false.  Can be negated with +b for short form,
// and --flag=false, or --flag=F for long form.  This is to facilitate
// shell scripts generating sets of arguments since defaults can be over-written
type Flag struct {
	//The long name of the flag, e.g., --force
	Long	string
	//Help string
	Help	string
	//Short option
	Short	byte
	//Whether flag was passed
	Passed	bool
}

//Create a new command flag
func NewFlag(short byte, long string, help string) *Flag {
	f := Flag{
		Long:	long,
		Short:	short,
		Help:	help,
	}
	flags = append(flags, f)
	optByShort[short] = &f
	optByLong[long] = &f
	return &f
}

//Creates a command argument that takes a single argument,
//that is over-written each time the flag is set.  E.g.,
//--foo=bar --foo=baz results in the "foo" flag having
//the value "baz"
type OptArg struct {
	Long	string
	Help	string
	Short	byte
	Opt	string
}

//Create a new OptArg
func NewOptArg(short byte, long string, help string) *OptArg {
	o := OptArg{
		Long:	long,
		Short:	short,
		Help:	help,
	}
	optArgs = append(optArgs, o)
	optByShort[short] = &o
	optByLong[long] = &o
	return &o
}

//Creates a command argument that can hold an array of arguments.  Each
//time the option appears, its argument is appended.
//E.g., --foo=bar --foo=baz results in "foo" having an array holding
//"bar" and "baz".
type OptVec struct {
	Long	string
	Help	string
	Short	byte
	OptArgs	[]string
}

//Construct a new OptVec
func NewOptVec(short byte, long string, help string) *OptVec {
	v := OptVec{
		Long:	long,
		Short:	short,
		Help:	help,
	}
	optVecs = append(optVecs, v)
	optByShort[short] = &v
	optByLong[long] = &v
	return &v
}

//An OptCount is like a flag, but holds the number of times it
//has been passed, minus the number of times it has been negated.
//You can also set the value directly.
//For example, where 'v' is the short option, "verbose" the long,
//"-vvv", "--verbose --verbose --verbose", and "--verbose=3" will
//all set that option to have a value of three.  The number is
//parsed with strconv.ParseInt with a base of 0, so binary, octal
//and hexadecimal numbers are parsed as well
//
//Intended for verbosity, debug level, etc.
type OptCount struct {
	Long	string
	Help	string
	Short	byte
	Count	int64
}

//Create new OptCount
func NewOptCount(short byte, long string, help string) *OptCount {
	c := OptCount{
		Long:	long,
		Short:	short,
		Help:	help,
	}
	optCounts = append(optCounts, c)
	optByShort[short] = &c
	optByLong[long] = &c
	return &c
}

const initialCapacity = 0

//Map of bytes to their associated options.  Used for parsing
//short options
var optByShort map[byte]any = make(map[byte]any, initialCapacity)

//Map of strings to options, used to parse long options
var optByLong map[string]any = make(map[string]any, initialCapacity)

//List of flags created
var flags []Flag = make([]Flag, 0, initialCapacity)

//List of optArgs created
var optArgs []OptArg = make([]OptArg, 0, initialCapacity)

//List of optVecs created
var optVecs []OptVec = make([]OptVec, 0, initialCapacity)

//List of optCounts created
var optCounts []OptCount = make([]OptCount, 0, initialCapacity)

//All arguments that were not program options
var Rest []string = make([]string, 0, initialCapacity)

//Current program version, used for printing version information
var ProgramVersion string

//Program name, if different from argv[0]
var ProgramName string

//Description of the program
var ProgramDesc string

//Call a function to process the argument '-'.  Normally,
//this will involve reading and processing data from standard
//input
var StdinHandler = func() error { return nil }

//Convert the strings "true", "false", "t", and "f" to
//their appropriate boolean values, case-insensitively,
//or return an error if some other string is passed
func optargToBool(s string) (bool, error) {
	if strings.EqualFold(s, "t") { return true, nil }
	if strings.EqualFold(s, "f") { return false, nil }
	if strings.EqualFold(s, "true") { return true, nil }
	if strings.EqualFold(s, "false") { return false, nil }
	return false, errors.New("Unable to parse boolean string passed as argument")
}

//Parse an array of strings as options
func ParseArgv(argv []string) error {
	expecting_optarg := false

	var waiting_opt *OptArg
	var waiting_vec *OptVec
	expecting_opt := false

	for i, arg := range argv {
		if len(arg) == 0 { continue }	//Skip empty arguments

		if expecting_opt {
			if expecting_optarg {
				waiting_opt.Opt = arg
			} else {
				waiting_vec.OptArgs = append(waiting_vec.OptArgs, arg)
			}
			expecting_opt = false
			continue
		}

		if len(arg) == 1 {
			if arg[0] == '-' {
				if e := StdinHandler(); e != nil {
					return e
				}
			} else {
				Rest = append(Rest, arg)
			}
			continue
		} else if len(arg) == 2 {
			if arg[0] == '-' {
				if arg[1] == '-' {
					for j := i + 1; j < len(argv); j++{
						Rest = append(Rest, argv[j])
					}
					return nil
				} else {
					if v, ok := optByShort[arg[1]]; ok {
						switch v.(type) {
						case *Flag:
							f := v.(*Flag)
							f.Passed = true
						case *OptArg:
							waiting_opt = v.(*OptArg)
							expecting_opt = true
							expecting_optarg = true
						case *OptVec:
							waiting_vec = v.(*OptVec)
							expecting_opt = true
							expecting_optarg = false
						case *OptCount:
							c := v.(*OptCount)
							c.Count++
						default:
							panic("Invalid flag type")
						}
					}
				}
			} else if arg[0] == '+' {
				if v, ok := optByShort[arg[1]]; ok {
					switch v.(type) {
					case *Flag:
						f := v.(*Flag)
						f.Passed = false
					case *OptArg:
						v.(*OptArg).Opt = ""
					case *OptVec:
						v.(*OptVec).OptArgs = make([]string, initialCapacity)
					case *OptCount:
						v.(*OptCount).Count--
					default:
						panic("Invalid flag type")
					}
				}
			} else {
				Rest = append(Rest, arg)
			}
		} else { //3 or more bytes
			if arg[0] == '-' {
				if arg[1] == '-' {	//Long argument
					equals := strings.IndexByte(arg, '=')
					if equals == -1 {
						if v, ok := optByLong[arg[2:]]; ok {
							switch v.(type) {
							case *Flag:
								f := v.(*Flag)
								f.Passed = true
							case *OptArg:
								waiting_opt = v.(*OptArg)
								expecting_opt = true
								expecting_optarg = true
							case *OptVec:
								waiting_vec = v.(*OptVec)
								expecting_opt = true
								expecting_optarg = false
							case *OptCount:
								c := v.(*OptCount)
								c.Count++
							default:
								panic("Invalid flag type")
							}
						} else {
							return errors.New(fmt.Sprintf("Unrecognized long option %s", arg[2:]))
						}
					} else {
						if v, ok := optByLong[arg[2:equals]]; ok {
							switch v.(type) {
							case *Flag:
								f := v.(*Flag)
								opt := arg[equals + 1:]
								val, err := optargToBool(opt)
								if err != nil {
									return err
								} else {
									f.Passed = val
								}
							case *OptArg:
								o := v.(*OptArg)
								opt := arg[equals + 1:]
								o.Opt = opt
							case *OptVec:
								o := v.(*OptVec)
								opt := arg[equals + 1:]
								o.OptArgs = append(o.OptArgs, opt)
							case *OptCount:
								if value, err := strconv.ParseInt(arg[equals + 1:], 0, 32); err != nil {
									return fmt.Errorf("Unable to parse %s as a number, %s", arg[equals + 1:], arg[2:equals])
								} else {
									v.(*OptCount).Count = value
								}
							default:
								panic("Invalid flag type")
							}
						}
					}
				} else {		//group of shorts
					for i := 1; i < len(arg); i++ {
						if v, ok := optByShort[arg[i]]; ok {
							switch v.(type) {
							case *Flag:
								f := v.(*Flag)
								f.Passed = true
							case *OptArg:
								o := v.(*OptArg)
								if i < len(arg) - 1 {
									o.Opt = arg[i + 1:]
									goto arg_loop_end
								} else {
									expecting_opt = true
									expecting_optarg = true
								}
							case *OptVec:
								o := v.(*OptVec)
								if i < len(arg) - 1 {
									o.OptArgs = append(o.OptArgs, arg[i + 1:])
									goto arg_loop_end
								} else {
									expecting_opt = true
									expecting_optarg = false
								}
							case *OptCount:
								c := v.(*OptCount)
								c.Count++
							default:
								panic("Invalid flag type")
							}
						} else {	//Invalid argument
							return fmt.Errorf("Unrecognized short option:  '%c'", arg[i])
						}
					}
					arg_loop_end:
				}
			} else if arg[0] == '+' {
				for i := 1; i < len(arg); i++ {
					if v, ok := optByShort[arg[i]]; ok {
						switch v.(type) {
						case *Flag:
							f := v.(*Flag)
							f.Passed = false
						case *OptArg:
							o := v.(*OptArg)
							o.Opt = ""
						case *OptVec:
							o := v.(*OptVec)
							o.OptArgs = make([]string, initialCapacity)
						case *OptCount:
							c := v.(*OptCount)
							c.Count--
						default:
							panic("Invalid flag type")
						}
					} else {	//Invalid argument
						return fmt.Errorf("Unrecognized short option:  '%c'", arg[i])
					}
				}
			} else {	//Not an option
				Rest = append(Rest, arg)
			}
		}
	}
	if expecting_opt {
		f := "Expecting argument for option:  -%c/--%s"
		if expecting_optarg {
			return fmt.Errorf(f, waiting_opt.Short, waiting_opt.Long)
		} else {
			return fmt.Errorf(f, waiting_vec.Short, waiting_vec.Long)
		}
	} else {
		return nil
	}
}

func GetOpts() error {
	if ProgramName == "" {
		ProgramName = os.Args[0]
	}
	if ProgramVersion == "" {
		ProgramVersion = "0.0.1"
	}
	return ParseArgv(os.Args[1:])
}

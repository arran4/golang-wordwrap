package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/arran4/golang-wordwrap/cmd/wordwrap/templates"
)

type Cmd interface {
	Execute(args []string) error
	Usage()
}

type InternalCommand struct {
	Exec      func(args []string) error
	UsageFunc func()
}

func (c *InternalCommand) Execute(args []string) error {
	return c.Exec(args)
}

func (c *InternalCommand) Usage() {
	c.UsageFunc()
}

type UserError struct {
	Err error
	Msg string
}

func (e *UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func NewUserError(err error, msg string) *UserError {
	return &UserError{Err: err, Msg: msg}
}

func executeUsage(out io.Writer, templateName string, data interface{}) error {
	return templates.GetTemplates().ExecuteTemplate(out, templateName, data)
}

type RootCmd struct {
	*flag.FlagSet
	Commands map[string]Cmd
	Version  string
	Commit   string
	Date     string
}

func (c *RootCmd) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	c.FlagSet.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
	for name := range c.Commands {
		fmt.Fprintf(os.Stderr, "    %s\n", name)
	}
}

func (c *RootCmd) UsageRecursive() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	c.FlagSet.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
	fmt.Fprintf(os.Stderr, "    %s\n", "dogoal")
	fmt.Fprintf(os.Stderr, "    %s\n", "pagelimits")
	fmt.Fprintf(os.Stderr, "    %s\n", "rich")
	fmt.Fprintf(os.Stderr, "    %s\n", "sample")
	fmt.Fprintf(os.Stderr, "    %s\n", "simple")
	fmt.Fprintf(os.Stderr, "    %s\n", "texttoimage")
}

func NewRoot(name, version, commit, date string) (*RootCmd, error) {
	c := &RootCmd{
		FlagSet:  flag.NewFlagSet(name, flag.ExitOnError),
		Commands: make(map[string]Cmd),
		Version:  version,
		Commit:   commit,
		Date:     date,
	}
	c.FlagSet.Usage = c.Usage
	c.Commands["dogoal"] = c.NewDogoal()
	c.Commands["pagelimits"] = c.NewPagelimits()
	c.Commands["rich"] = c.NewRich()
	c.Commands["sample"] = c.NewSample()
	c.Commands["simple"] = c.NewSimple()
	c.Commands["texttoimage"] = c.NewTexttoimage()
	c.Commands["help"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					c.UsageRecursive()
					return nil
				}
			}
			c.Usage()
			return nil
		},
		UsageFunc: c.Usage,
	}
	c.Commands["usage"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					c.UsageRecursive()
					return nil
				}
			}
			c.Usage()
			return nil
		},
		UsageFunc: c.Usage,
	}
	c.Commands["version"] = &InternalCommand{
		Exec: func(args []string) error {
			fmt.Printf("Version: %s\nCommit: %s\nDate: %s\n", c.Version, c.Commit, c.Date)
			return nil
		},
		UsageFunc: func() {
			fmt.Fprintf(os.Stderr, "Usage: %s version\n", os.Args[0])
		},
	}
	return c, nil
}

func (c *RootCmd) Execute(args []string) error {
	if err := c.FlagSet.Parse(args); err != nil {
		return NewUserError(err, fmt.Sprintf("flag parse error %s", err.Error()))
	}
	remainingArgs := c.FlagSet.Args()
	if len(remainingArgs) < 1 {
		c.Usage()
		return nil
	}
	cmd, ok := c.Commands[remainingArgs[0]]
	if !ok {
		c.Usage()
		return fmt.Errorf("unknown command: %s", remainingArgs[0])
	}
	return cmd.Execute(remainingArgs[1:])
}

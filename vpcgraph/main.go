package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"os"
	"os/exec"

	"io"
	"log"
	"strings"
)

type Printable interface {
	Print(int, io.Writer)
}

type Graph struct {
	Name     string
	Label    string
	Children []Printable
}

func (self *Graph) Print(depth int, out io.Writer) {
	prefix := Indent(depth)
	fmt.Fprintf(out, "%sgraph %s {\n", prefix, self.Name)

	for _, child := range self.Children {
		child.Print(depth+1, out)
	}
	prefix = Indent(depth + 1)
	fmt.Fprintf(out, "%slabel=\"%s\";\n", prefix, self.Label)
	prefix = Indent(depth)
	fmt.Fprintf(out, "%s}\n", prefix)
}

type Subgraph struct {
	Name     string
	Children []Printable
	Label    string
	Nodes    []string
}

func (self Subgraph) Print(depth int, out io.Writer) {
	prefix := Indent(depth)
	fmt.Fprintf(out, "%ssubgraph cluster_%s {\n", prefix, self.Name)
	prefix = Indent(depth + 1)
	fmt.Fprintf(out, "%scolor=lightgrey;\n", prefix)
	fmt.Fprintf(out, "%snode [style=filled,color=lightgrey];\n", prefix)

	fmt.Fprintf(out, "%slabel = \"%s\";\n", prefix, self.Label)

	for _, node := range self.Nodes {
		fmt.Fprintf(out, "%s%s;\n", prefix, node)
	}
	prefix = Indent(depth)
	fmt.Fprintf(out, "%s}\n", prefix)
}

func FindTag(name string, tags []*ec2.Tag) *ec2.Tag {
	for _, tag := range tags {
		if *tag.Key == name {
			return tag
		}
	}
	return &ec2.Tag{Value: aws.String("")}
}

func FindNameTag(tags []*ec2.Tag) string {
	tag := FindTag("Name", tags)
	return DotSafe(*tag.Value)
}

func DotSafe(str string) string {
	return strings.Replace(str, "-", "_", -1)
}

func Indent(depth int) string {
	depth = depth * 2
	s := ""
	for i := 0; i < depth; i++ {
		s += " "
	}
	return s
}

func main() {
	fmt.Println("Running")
	svc := ec2.New(&aws.Config{Region: aws.String("us-west-2")})
	fmt.Println("Created EC@ Client")

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create("vpc.dot")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	vpc := resp.Vpcs[0]
	fmt.Println(vpc)

	vpcName := FindNameTag(vpc.Tags)
	fmt.Println(vpcName)

	graph := Graph{
		Name: "DiagramName",
	}

	subnet := Subgraph{
		Name:  vpcName,
		Label: fmt.Sprintf("%s: %s", vpcName, DotSafe(*vpc.CidrBlock)),
		Nodes: []string{"_"},
	}

	graph.Children = append(graph.Children, subnet)

	graph.Print(0, file)

	exec.Command("dot", "-Tpng", "-o", "vpc.png", "vpc.dot").Start()
	exec.Command("open", "vpc.png").Start()
}

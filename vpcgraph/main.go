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
	Print(io.Writer)
}

type Graph struct {
	Name     string
	Children []Printable
}

func (self *Graph) Print(out io.Writer) {
	fmt.Fprintf(out, "graph %s {\n", self.Name)

	for _, child := range self.Children {
		child.Print(out)
	}
	fmt.Fprintf(out, "label=\"%s\";\n", self.Name)
	fmt.Fprint(out, "}\n")
}

type Subgraph struct {
	Name     string
	Children []Printable
	Nodes    []string
}

func (self Subgraph) Print(out io.Writer) {
	fmt.Fprintf(out, "subgraph cluster_%s {\n", self.Name)
	fmt.Fprintf(out, "color=lightgrey;\n")
	fmt.Fprintf(out, "node [style=filled,color=lightgrey];\n")

	fmt.Fprintf(out, "label = \"%s\";\n", self.Name)

	for _, node := range self.Nodes {
		fmt.Fprintf(out, "%s;\n", node)
	}

	fmt.Fprint(out, "}\n")
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

	vpcName := FindNameTag(vpc.Tags)
	fmt.Println(vpcName)

	graph := Graph{
		Name: "DiagramName",
	}

	subnet := Subgraph{
		Name:  vpcName,
		Nodes: []string{"a1", "a2"},
	}

	graph.Children = append(graph.Children, subnet)

	graph.Print(file)

	exec.Command("dot", "-Tpng", "-o", "vpc.png", "vpc.dot").Start()
	exec.Command("open", "vpc.png").Start()
}

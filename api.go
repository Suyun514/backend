package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func APIMetadata(c *gin.Context) {
	c.JSON(200, gin.H{
		"static": false,
	})
}

func APIProblemSetList(c *gin.Context) {
	c.JSON(200, &ProblemSetFile)
}

func APIProblemSetMetadata(c *gin.Context) {
	problemset := c.Param("problemset")
	for _, i := range ProblemSets {
		if i.Id == problemset {
			c.JSON(200, gin.H{
				"name":    i.Name,
				"problem": len(i.ProblemArray),
				"page":    i.MaxPage,
			})
			return
		}
	}
	c.JSON(404, gin.H{
		"errorCode":    404,
		"errorMessage": "Problem set not found.",
	})
}

func APIProblemSetPage(c *gin.Context) {
	problemset := c.Param("problemset")
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		c.JSON(404, gin.H{
			"errorCode":    404,
			"errorMessage": "page invalid",
		})
		return
	}

	for _, i := range ProblemSets {
		if i.Id == problemset {
			if page < 0 || page > i.MaxPage {
				c.JSON(404, gin.H{
					"errorCode":    404,
					"errorMessage": "page invalid",
				})
				return
			}
			c.JSON(200, i.ProblemArray[50*page-49:min(50*page, len(i.ProblemArray))])
			return
		}
	}
	c.JSON(404, gin.H{
		"errorCode":    404,
		"errorMessage": "Problem set not found.",
	})
}

func APIProblem(c *gin.Context) {
	problemset := c.Param("problemset")
	problem := c.Param("problem")
	for _, i := range ProblemSets {
		if i.Id == problemset {
			if _, ok := i.ProblemMap[problem]; !ok {
				c.JSON(404, gin.H{
					"errorCode":    404,
					"errorMessage": "Problem not found",
				})
				return
			}
			b, err := ioutil.ReadFile(BasePath + "/" + i.Id + "/" + problem + "/main.json")
			if err != nil {
				ThrowUnknownError(c, fmt.Errorf("error can't read file %s/%s/main.json : [%w] ", i.Id, problem, err))
				return
			}
			var res interface{}
			err = json.Unmarshal(b, &res)
			if err != nil {
				ThrowUnknownError(c, fmt.Errorf("error can't parse file %s/%s/main.json : [%w] ", i.Id, problem, err))
				return
			}
			descBytes, err := ioutil.ReadFile(BasePath + "/" + i.Id + "/" + problem + "/description.md")
			if err != nil {
				ThrowUnknownError(c, fmt.Errorf("error can't read file %s/%s/description.md : [%w] ", i.Id, problem, err))
				return
			}
			desc := string(descBytes)
			if res, ok := res.(map[string]interface{}); ok {
				descType, ok := res["description_type"]
				if !ok || descType == "" {
					descType = "markdown"
				} else {
					delete(res, "description_type")
				}
				if descType == "html" {
					lines := strings.Split(desc, "\n")
					flag := false
					output := ""
					if len(lines) > 0 && len(lines[0]) > 0 && lines[0][0] != '#' {
						flag = true
						output += `<div class="oiarchive-block">`
					}
					for _, j := range lines {
						if len(j) == 0 {
							continue
						}
						if len(j) < 2 {
							output += j + "\n"
							continue
						}
						jj := strings.Split(j, " ")
						if jj[0] == "#" {
							if flag {
								output += `</div>` + "\n"
							}
							flag = true
							output += `<div class="oiarchive-block">` + "\n"
							output += `<h4 class="oiarchive-block-title">` + j[2:] + `</h4>` + "\n"
						} else {
							output += j + "\n"
						}
					}
					if flag {
						output += `</div>` + "\n"
					}
					desc = output
				}
				if descType == "markdown" {
					res["description"] = string(blackfriday.Run([]byte(desc)))
				} else {
					res["description"] = desc
				}
				log.Println(desc)
				log.Println(res["description"])
				c.JSON(200, res)
			} else {
				ThrowUnknownError(c, fmt.Errorf("error can't parse file %s/%s/main.json", i.Id, problem))
				return
			}
			return
		}
	}
	c.JSON(404, gin.H{
		"errorCode":    404,
		"errorMessage": "Problem set not found.",
	})
}

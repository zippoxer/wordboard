from dataclasses import dataclass
from typing import List
import random
from tabulate import tabulate


@dataclass
class Point:
    x: int
    y: int


@dataclass
class Rect:
    min: Point
    max: Point


class Board:
    width: int
    height: int

    def __init__(self, width: int, height: int):
        self.width = width
        self.height = height
        self.clear()

    def clear(self):
        self.chars = [[None for x in range(self.height)]
                      for y in range(self.width)]

    # Scans an area for instances of char
    # and returns a list of their locations.
    def scan(self, char: str, area: Rect = None) -> List[Point]:
        if not area:
            area = Rect(
                Point(0, 0),
                Point(self.width-1, self.height-1)
            )

        instances = []
        for x in range(area.min.x, area.max.x+1):
            for y in range(area.min.y, area.max.y+1):
                if self.chars[x][y] == char:
                    instances.append(Point(x, y))
        return instances

    def scan_adjacent(self, char: str, slot: Point) -> List[Point]:
        area = Rect(
            Point(
                max(slot.x - 1, 0),
                max(slot.y - 1, 0)
            ),
            Point(
                min(slot.x + 1, self.width - 1),
                min(slot.y + 1, self.height - 1)
            )
        )

        instances = []
        for x in range(area.min.x, area.max.x+1):
            for y in range(area.min.y, area.max.y+1):
                if self.chars[x][y] == char and not (slot.x == x and slot.y == y):
                    instances.append(Point(x, y))
        return instances


class Filler:
    board: Board

    def __init__(self, board: Board):
        self.board = board

    def fill(self, words: List[str]) -> List[List[Point]]:
        paths = []
        for word in words:
            path = self.__insert_word(word)
            if not path:
                return None

            paths.append(path)

        return paths

    def __insert_word(self, word: str) -> List[Point]:
        # Find a random empty slot to insert the word at.
        empty_slots = self.board.scan(None)
        if not empty_slots:
            return False

        slot = random.choice(empty_slots)

        path = []
        for char in word:
            self.board.chars[slot.x][slot.y] = char
            path.append(slot)

            # Find all adjacent empty slots and choose
            # a random one to insert the next letter at.
            empty_slots = self.board.scan_adjacent(None, slot)
            if not empty_slots:
                return None

            slot = random.choice(empty_slots)

        return path


class Validator:
    def __init__(self, board: Board):
        self.board = board

    def validate(self, words: List[str], paths: List[List[Point]]) -> bool:
        # Find all possible starting locations of each word,
        # and check each location to determine whether
        # the word can be selected in any other path
        # than the one that's been defined for it.
        for i, word in enumerate(words):
            for slot in self.board.scan(word[0]):
                if self.__check_collision(word[1:], paths[i], slot):
                    return False

        return True

    def __check_collision(self, word: str, path: List[Point], pos: Point, out_of_path: bool = False) -> bool:
        if len(word) == 0:
            return out_of_path

        for slot in self.board.scan_adjacent(word[0], pos):
            if slot not in path:
                out_of_path = True

            if self.__check_collision(word[1:], path, slot, out_of_path):
                return True

        return False


board = Board(5, 5)
words = ["מה", "מצב", "מותק", "איך", "אתה", "מרגיש"]

while True:
    board.clear()

    filler = Filler(board)
    paths = filler.fill(words)
    if not paths:
        continue

    validator = Validator(board)
    if not validator.validate(words, paths):
        continue

    break

print(tabulate(board.chars, tablefmt="grid"))
